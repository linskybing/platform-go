package configfile

import (
	"fmt"
	"strings"
)

// findPodSpecs recursively looks for objects that look like PodSpecs.
// It searches for "containers" array or specific nesting keys.
func findPodSpecs(obj map[string]interface{}) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, 4) // Pre-allocate for typical K8s object (1-2 specs)

	// Direct match: strict check for containers
	if _, hasContainers := obj["containers"]; hasContainers {
		return append(results, obj)
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
	// Pre-allocate for typical pod (1-3 containers + 0-2 init containers)
	containers := make([]map[string]interface{}, 0, 5)

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
