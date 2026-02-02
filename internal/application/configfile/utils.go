package configfile

import (
	"fmt"
	"strings"
)

// findPodSpecs recursively looks for objects that look like PodSpecs.
// It searches for "containers" array or specific nesting keys.
func findPodSpecs(obj map[string]interface{}) []map[string]interface{} {
	var results []map[string]interface{}

	// Direct match: strict check for containers
	if _, hasContainers := obj["containers"]; hasContainers {
		results = append(results, obj)
		// Return here? No, theoretically a PodSpec could define a sidecar that manages other pods (rare but possible in CRDs)
		// But for standard K8s objects (Pod, Deployment, Job), the PodSpec is a leaf regarding structure.
		// We continue traversal only if strictly necessary.
		// For performance, assuming standard K8s structure, we can stop deeper traversal for this branch.
		return results
	}

	// Recursive Search
	for key, value := range obj {
		if subMap, ok := value.(map[string]interface{}); ok {
			// Optimized: Only traverse relevant keys for K8s controllers
			if key == "spec" || key == "template" || key == "jobTemplate" {
				results = append(results, findPodSpecs(subMap)...)
			}
		}
	}
	return results
}

func getContainersFromPodSpec(podSpec map[string]interface{}) []map[string]interface{} {
	var containers []map[string]interface{}

	appendContainers := func(key string) {
		if list, ok := podSpec[key].([]interface{}); ok {
			for _, item := range list {
				if c, ok := item.(map[string]interface{}); ok {
					containers = append(containers, c)
				}
			}
		}
	}

	appendContainers("containers")
	appendContainers("initContainers")
	return containers
}

func parseImageNameTag(img string) (name string, tag string) {
	lastColon := strings.LastIndex(img, ":")
	lastSlash := strings.LastIndex(img, "/")

	if lastColon == -1 || lastColon < lastSlash {
		return img, "latest"
	}

	return img[:lastColon], img[lastColon+1:]
}

func getStringValue(m map[string]interface{}, key string) (string, bool) {
	val, ok := m[key]
	if !ok {
		return "", false
	}

	switch v := val.(type) {
	case string:
		return v, true
	case float64:
		return fmt.Sprintf("%g", v), true
	case int, int64, int32:
		return fmt.Sprintf("%d", v), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}
