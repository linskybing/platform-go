package k8s

import ()

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// WatchNamespaceResources monitors resources for a specific namespace
func WatchNamespaceResources(ctx context.Context, writeChan chan<- []byte, namespace string) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	var wg sync.WaitGroup
	for _, gvr := range gvrs {
		wg.Add(1)
		time.Sleep(50 * time.Millisecond) // Stagger start to be gentle on APIServer

		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchAndSend(ctx, DynamicClient, gvr, namespace, writeChan)
		}(gvr)
	}

	// Wait for all watchers to finish (via context cancel) then close channel
	go func() {
		<-ctx.Done()
		wg.Wait()
		close(writeChan)
	}()
}

// WatchNamespaceResources monitors resource changes in a single namespace
func WatchUserNamespaceResources(ctx context.Context, namespace string, writeChan chan<- []byte) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	var wg sync.WaitGroup
	for _, gvr := range gvrs {
		wg.Add(1)
		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchUserAndSend(ctx, namespace, gvr, writeChan)
		}(gvr)
	}

	// Wait logic handled by caller or context
	wg.Wait()
}

func watchUserAndSend(ctx context.Context, namespace string, gvr schema.GroupVersionResource, writeChan chan<- []byte) {
	// lastSnapshot holds last sent status signature per resource name
	lastSnapshot := make(map[string]string)

	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		name := obj.GetName()

		// Always send deletes
		if eventType != "DELETED" {
			// Compute compact snapshot to decide whether to send
			snap := statusSnapshotString(obj)
			if prev, ok := lastSnapshot[name]; ok {
				if prev == snap {
					// No meaningful status change, skip sending
					return nil
				}
			}
			lastSnapshot[name] = snap
		} else {
			// remove snapshot on delete
			delete(lastSnapshot, name)
		}

		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Prevent blocking if client is slow
			return fmt.Errorf("client buffer full, dropping message for %s", name)
		}
	}

	list, err := DynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			_ = sendObject("ADDED", &item)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 30):
			// Simple reconnection logic
			watcher, err := DynamicClient.Resource(gvr).Namespace(namespace).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			func() {
				defer watcher.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case event, ok := <-watcher.ResultChan():
						if !ok {
							return
						}
						if obj, ok := event.Object.(*unstructured.Unstructured); ok {
							_ = sendObject(string(event.Type), obj)
						}
					}
				}
			}()
		}
	}
}

func watchAndSend(
	ctx context.Context,
	dynClient dynamic.Interface,
	gvr schema.GroupVersionResource,
	ns string,
	writeChan chan<- []byte,
) {
	// lastSnapshot holds last sent status signature per resource name
	lastSnapshot := make(map[string]string)

	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		name := obj.GetName()

		if eventType != "DELETED" {
			snap := statusSnapshotString(obj)
			if prev, ok := lastSnapshot[name]; ok {
				if prev == snap {
					return nil
				}
			}
			lastSnapshot[name] = snap
		} else {
			delete(lastSnapshot, name)
		}

		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Initial List
	list, err := dynClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			if err := sendObject("ADDED", &item); err != nil && ctx.Err() != context.Canceled {
				slog.Warn("Failed to send list item", slog.Any("error", err))
			}
		}
	} else if ctx.Err() == context.Canceled {
		return
	} else {
		slog.Error("List error for resource",
			slog.String("resource", gvr.Resource),
			slog.String("group", gvr.Group),
			slog.Any("error", err))
	}

	// Watch Loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := dynClient.Resource(gvr).Namespace(ns).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			if ctx.Err() == context.Canceled {
				return
			}
			time.Sleep(5 * time.Second)
			continue
		}

		func() {
			defer watcher.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-watcher.ResultChan():
					if !ok {
						return
					}

					obj, ok := event.Object.(*unstructured.Unstructured)
					if !ok {
						continue
					}

					if err := sendObject(string(event.Type), obj); err != nil && ctx.Err() != context.Canceled {
						slog.Warn("Failed to send watch event", slog.Any("error", err))
					}
				}
			}
		}()
	}
}

// buildDataMap extracts comprehensive details from K8s resources for the frontend
func buildDataMap(eventType string, obj *unstructured.Unstructured) map[string]interface{} {
	data := map[string]interface{}{
		"type": eventType,
		"kind": obj.GetKind(),
		"name": obj.GetName(),
		"ns":   obj.GetNamespace(),
	}

	metadata := map[string]interface{}{}
	if ts, found, _ := unstructured.NestedString(obj.Object, "metadata", "creationTimestamp"); found {
		metadata["creationTimestamp"] = ts
	}
	if labels, found, _ := unstructured.NestedStringMap(obj.Object, "metadata", "labels"); found {
		metadata["labels"] = labels
	}
	if dts, found, _ := unstructured.NestedString(obj.Object, "metadata", "deletionTimestamp"); found {
		metadata["deletionTimestamp"] = dts
	}
	data["metadata"] = metadata

	for k, v := range extractStatusFields(obj) {
		data[k] = v
	}

	// Provide helpful lifecycle hints for the frontend when events indicate creation/termination
	// - DELETED: ensure deletionTimestamp exists and mark status as Terminating so UI can show it
	// - ADDED for Pods: if phase is Pending or no status yet, mark as Creating
	if eventType == "DELETED" {
		if _, ok := metadata["deletionTimestamp"]; !ok {
			metadata["deletionTimestamp"] = time.Now().Format(time.RFC3339)
			data["metadata"] = metadata
		}
		// Only set status if not already set by extractStatusFields
		if _, exists := data["status"]; !exists {
			data["status"] = "Terminating"
		}
	}

	if eventType == "ADDED" && obj.GetKind() == "Pod" {
		// If status/phase is missing or Pending, surface a Creating hint
		if s, ok := data["status"].(string); !ok || strings.EqualFold(s, "Pending") || s == "" {
			data["status"] = "Creating"
		}
	}

	if obj.GetKind() == "Pod" {
		if containers, found, _ := unstructured.NestedSlice(obj.Object, "spec", "containers"); found {
			var containerNames []string
			var images []string
			for _, c := range containers {
				if m, ok := c.(map[string]interface{}); ok {
					if name, ok := m["name"].(string); ok {
						containerNames = append(containerNames, name)
					}
					if image, ok := m["image"].(string); ok {
						images = append(images, image)
					}
				}
			}
			if len(containerNames) > 0 {
				data["containers"] = containerNames
			}
			if len(images) > 0 {
				data["images"] = images
			}
		}

		if containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses"); found {
			var totalRestarts int64 = 0
			for _, cs := range containerStatuses {
				if m, ok := cs.(map[string]interface{}); ok {
					if rc, ok := m["restartCount"].(int64); ok {
						totalRestarts += rc
					}
				}
			}
			data["restartCount"] = totalRestarts
		}
	}

	// Attach recent events for Pods to help frontend show `kubectl describe` style information
	if obj.GetKind() == "Pod" {
		if evs := fetchPodEvents(obj.GetNamespace(), obj.GetName()); len(evs) > 0 {
			data["events"] = evs
		}
	}

	if isService(obj) {
		if ips := extractServiceExternalIPs(obj); len(ips) > 0 {
			data["externalIPs"] = ips
		}
		if ports := extractServiceNodePorts(obj); len(ports) > 0 {
			data["nodePorts"] = ports
		}
		if ports := extractServicePorts(obj); len(ports) > 0 {
			data["ports"] = ports
		}
	}

	return data
}

func extractServicePorts(obj *unstructured.Unstructured) []string {
	var servicePorts []string
	ports, found, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
	if !found || err != nil {
		return servicePorts
	}

	for _, port := range ports {
		if m, ok := port.(map[string]interface{}); ok {
			p, okPort := m["port"].(int64)
			proto, okProto := m["protocol"].(string)

			if okPort {
				portStr := fmt.Sprintf("%d", p)
				if okProto {
					portStr = fmt.Sprintf("%d/%s", p, proto)
				}
				servicePorts = append(servicePorts, portStr)
			}
		}
	}
	return servicePorts
}

func isService(obj *unstructured.Unstructured) bool {
	return obj.GetKind() == "Service"
}

func extractServiceExternalIPs(obj *unstructured.Unstructured) []string {
	var externalIPs []string

	specExternalIPs, found, err := unstructured.NestedSlice(obj.Object, "spec", "externalIPs")
	if found && err == nil {
		for _, ip := range specExternalIPs {
			if s, ok := ip.(string); ok {
				externalIPs = append(externalIPs, s)
			}
		}
	}

	ingressList, found, err := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress")
	if found && err == nil {
		for _, ingress := range ingressList {
			if m, ok := ingress.(map[string]interface{}); ok {
				if ip, ok := m["ip"].(string); ok {
					externalIPs = append(externalIPs, ip)
				}
			}
		}
	}

	return externalIPs
}

func extractServiceNodePorts(obj *unstructured.Unstructured) []int64 {
	var nodePorts []int64

	ports, found, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
	if !found || err != nil {
		return nodePorts
	}

	for _, port := range ports {
		if m, ok := port.(map[string]interface{}); ok {
			if np, ok := m["nodePort"].(int64); ok {
				nodePorts = append(nodePorts, np)
			} else if npf, ok := m["nodePort"].(float64); ok {
				nodePorts = append(nodePorts, int64(npf))
			}
		}
	}

	return nodePorts
}

func extractStatusFields(obj *unstructured.Unstructured) map[string]interface{} {
	kind := obj.GetKind()
	result := map[string]interface{}{}

	switch kind {
	case "Pod":
		if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
			result["status"] = phase
		}

		// Detect CrashLoopBackOff by inspecting containerStatuses.state.waiting.reason
		if containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses"); found {
			var crashContainers []string
			for _, cs := range containerStatuses {
				m, ok := cs.(map[string]interface{})
				if !ok {
					continue
				}

				// container name
				name := ""
				if n, ok := m["name"].(string); ok {
					name = n
				}

				if state, ok := m["state"].(map[string]interface{}); ok {
					if waiting, ok := state["waiting"].(map[string]interface{}); ok {
						if reason, ok := waiting["reason"].(string); ok {
							if strings.Contains(reason, "CrashLoopBackOff") {
								crashContainers = append(crashContainers, name)
								// prefer reporting CrashLoopBackOff as the pod status for UI clarity
								result["status"] = "CrashLoopBackOff"
								result["statusReason"] = reason
								break
							}
						}
						// also check message if reason not present
						if msg, ok := waiting["message"].(string); ok && strings.Contains(msg, "CrashLoopBackOff") {
							crashContainers = append(crashContainers, name)
							result["status"] = "CrashLoopBackOff"
							result["statusReason"] = msg
							break
						}
					}
				}
			}

			if len(crashContainers) > 0 {
				result["crashLoopContainers"] = crashContainers
			}
		}
	case "Service":
		if clusterIP, found, _ := unstructured.NestedString(obj.Object, "spec", "clusterIP"); found {
			result["clusterIP"] = clusterIP
		}
		if externalIPs, found, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); found && len(externalIPs) > 0 {
			if ingressMap, ok := externalIPs[0].(map[string]interface{}); ok {
				if ip, ok := ingressMap["ip"].(string); ok {
					result["externalIP"] = ip
				}
				if hostname, ok := ingressMap["hostname"].(string); ok {
					result["externalHostname"] = hostname
				}
			}
		}
	case "Ingress":
		if externalIPs, found, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); found && len(externalIPs) > 0 {
			if ingressMap, ok := externalIPs[0].(map[string]interface{}); ok {
				if ip, ok := ingressMap["ip"].(string); ok {
					result["externalIP"] = ip
				}
				if hostname, ok := ingressMap["hostname"].(string); ok {
					result["externalHostname"] = hostname
				}
			}
		}
	case "Deployment", "ReplicaSet":
		if availableReplicas, found, _ := unstructured.NestedInt64(obj.Object, "status", "availableReplicas"); found {
			result["availableReplicas"] = availableReplicas
		}
	case "Job":
		if succeeded, found, _ := unstructured.NestedInt64(obj.Object, "status", "succeeded"); found {
			result["succeeded"] = succeeded
		}
	}

	return result
}

// statusSnapshotString produces a compact, stable string representing the
// resource's status-related fields used for change detection. Keep this small
// to avoid expensive allocations; it's used to deduplicate frequent identical
// events (e.g. unrelated metadata updates).
func statusSnapshotString(obj *unstructured.Unstructured) string {
	m := map[string]interface{}{}

	// include kind/name for clarity (not strictly necessary for map key)
	m["kind"] = obj.GetKind()
	m["name"] = obj.GetName()

	// metadata.deletionTimestamp if present
	if dts, found, _ := unstructured.NestedString(obj.Object, "metadata", "deletionTimestamp"); found {
		m["deletionTimestamp"] = dts
	}

	// include extractStatusFields output (only status-related keys)
	for k, v := range extractStatusFields(obj) {
		m[k] = v
	}

	// For Pods, also include container restart counts and waiting reasons
	if obj.GetKind() == "Pod" {
		if containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses"); found {
			var csSnap []map[string]interface{}
			for _, cs := range containerStatuses {
				if cm, ok := cs.(map[string]interface{}); ok {
					entry := map[string]interface{}{}
					if name, ok := cm["name"].(string); ok {
						entry["name"] = name
					}
					if rc, ok := cm["restartCount"].(int64); ok {
						entry["restartCount"] = rc
					} else if rcf, ok := cm["restartCount"].(float64); ok {
						entry["restartCount"] = int64(rcf)
					}
					if state, ok := cm["state"].(map[string]interface{}); ok {
						if waiting, ok := state["waiting"].(map[string]interface{}); ok {
							if reason, ok := waiting["reason"].(string); ok {
								entry["waitingReason"] = reason
							}
							if msg, ok := waiting["message"].(string); ok {
								entry["waitingMessage"] = msg
							}
						}
					}
					csSnap = append(csSnap, entry)
				}
			}
			if len(csSnap) > 0 {
				m["containerStatuses"] = csSnap
			}
		}
	}

	// Marshal into compact JSON string for easy equality checks
	bs, _ := json.Marshal(m)
	return string(bs)
}

// fetchPodEvents retrieves recent events related to a Pod (namespace/name).
// Returns a compact serializable slice of event objects for frontend consumption.
func fetchPodEvents(namespace, name string) []map[string]interface{} {
	if Clientset == nil {
		return nil
	}

	// Field selector on involvedObject.name (events are namespaced)
	opts := metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", name)}
	list, err := Clientset.CoreV1().Events(namespace).List(context.TODO(), opts)
	if err != nil || list == nil {
		return nil
	}

	// Build a small array of events sorted by LastTimestamp ascending (older -> newer)
	var evs []map[string]interface{}
	for _, e := range list.Items {
		ev := map[string]interface{}{
			"type":           e.Type,
			"reason":         e.Reason,
			"message":        e.Message,
			"count":          e.Count,
			"firstTimestamp": e.FirstTimestamp.String(),
			"lastTimestamp":  e.LastTimestamp.String(),
			"source":         e.Source.Component,
		}
		evs = append(evs, ev)
	}

	// If there are many events, return the most recent 20
	if len(evs) > 20 {
		return evs[len(evs)-20:]
	}
	return evs
}
