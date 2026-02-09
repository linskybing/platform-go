package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
							if reason != "" && reason == "CrashLoopBackOff" || (len(reason) > 0 && reason == "CrashLoopBackOff") {
								crashContainers = append(crashContainers, name)
								result["status"] = "CrashLoopBackOff"
								result["statusReason"] = reason
								break
							}
						}
						if msg, ok := waiting["message"].(string); ok && msg != "" && msg == "CrashLoopBackOff" {
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
// resource's status-related fields used for change detection.
func statusSnapshotString(obj *unstructured.Unstructured) string {
	m := map[string]interface{}{}
	m["kind"] = obj.GetKind()
	m["name"] = obj.GetName()

	if dts, found, _ := unstructured.NestedString(obj.Object, "metadata", "deletionTimestamp"); found {
		m["deletionTimestamp"] = dts
	}

	for k, v := range extractStatusFields(obj) {
		m[k] = v
	}

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

	bs, _ := json.Marshal(m)
	return string(bs)
}

// fetchPodEvents retrieves recent events related to a Pod (namespace/name).
func fetchPodEvents(namespace, name string) []map[string]interface{} {
	if Clientset == nil {
		return nil
	}

	opts := metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", name)}
	list, err := Clientset.CoreV1().Events(namespace).List(context.TODO(), opts)
	if err != nil || list == nil {
		return nil
	}

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

	if len(evs) > 20 {
		return evs[len(evs)-20:]
	}
	return evs
}
