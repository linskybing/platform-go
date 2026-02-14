package k8s

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ToSafeK8sName(rawName string) string {
	safeName := strings.ToLower(rawName)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	safeName = reg.ReplaceAllString(safeName, "-")

	safeName = strings.Trim(safeName, "-")

	multiHyphenReg := regexp.MustCompile(`-+`)
	safeName = multiHyphenReg.ReplaceAllString(safeName, "-")

	if len(safeName) > 63 {
		safeName = safeName[:63]
		safeName = strings.TrimRight(safeName, "-")
	}

	if safeName == "" {
		safeName = "unnamed"
	}

	return safeName
}

// GenerateSafeResourceName generates a unique and K8s-compliant resource name.
func GenerateSafeResourceName(prefix string, name string, id string) string {
	reg := regexp.MustCompile("[^a-z0-9]+")
	safeName := reg.ReplaceAllString(strings.ToLower(name), "-")
	safeName = strings.Trim(safeName, "-")

	hashInput := fmt.Sprintf("project-%s", id)
	hash := sha256.Sum256([]byte(hashInput))
	shortHash := fmt.Sprintf("%x", hash)[:6]

	baseName := fmt.Sprintf("%s-%s", prefix, safeName)
	suffix := fmt.Sprintf("-%s", shortHash)

	maxLength := 63 - len(suffix)
	if len(baseName) > maxLength {
		baseName = baseName[:maxLength]
		baseName = strings.TrimRight(baseName, "-")
	}

	return baseName + suffix
}

func extractStatusFields(obj *unstructured.Unstructured) map[string]interface{} {
	kind := obj.GetKind()
	result := map[string]interface{}{}

	switch kind {
	case "Pod":
		if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
			result["status"] = phase
		}

		if containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses"); found {
			var crashContainers []string
			for _, cs := range containerStatuses {
				m, ok := cs.(map[string]interface{})
				if !ok {
					continue
				}
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

func statusSnapshotString(obj *unstructured.Unstructured) string {
	fields := extractStatusFields(obj)

	kind := obj.GetKind()
	switch kind {
	case "Pod":
		status, _ := fields["status"]
		crash, _ := fields["crashLoopContainers"]
		return fmt.Sprintf("Pod status:%v crash:%v", status, crash)
	case "Service":
		cIP, _ := fields["clusterIP"]
		extIP, _ := fields["externalIP"]
		return fmt.Sprintf("Service clusterIP:%v externalIP:%v", cIP, extIP)
	case "Deployment":
		avail, _ := fields["availableReplicas"]
		return fmt.Sprintf("Deployment available:%v", avail)
	default:
		return fmt.Sprintf("%v", fields)
	}
}

// fetchPodEvents retrieves recent events related to a Pod (namespace/name).
// Returns a compact serializable slice of event objects for frontend consumption.
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
