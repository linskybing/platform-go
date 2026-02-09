package k8s

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

	if eventType == "DELETED" {
		if _, ok := metadata["deletionTimestamp"]; !ok {
			metadata["deletionTimestamp"] = time.Now().Format(time.RFC3339)
			data["metadata"] = metadata
		}
		if _, exists := data["status"]; !exists {
			data["status"] = "Terminating"
		}
	}

	if eventType == "ADDED" && obj.GetKind() == "Pod" {
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
