package configfile

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/datatypes"
	k8sRes "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

// parseAndValidateResources splits raw YAML into documents, validates them, and prepares Resource structs.
func (s *ConfigFileService) parseAndValidateResources(rawYaml string) ([]*resource.Resource, error) {
	yamlArray := utils.SplitYAMLDocuments(rawYaml)
	if len(yamlArray) == 0 {
		return nil, ErrNoValidYAMLDocument
	}

	resourcesToCreate := make([]*resource.Resource, 0, len(yamlArray))

	for i, doc := range yamlArray {
		// Convert YAML to JSON
		jsonBytes, err := yaml.YAMLToJSON([]byte(doc))
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON for document %d: %w", i+1, err)
		}

		// Parse JSON to map for logical validation
		var obj map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &obj); err != nil {
			return nil, fmt.Errorf("failed to parse JSON for validation in document %d: %w", i+1, err)
		}

		// Validate Container Limits (Business Logic)
		if err := validateContainerLimits(obj); err != nil {
			return nil, fmt.Errorf("validation failed in document %d: %w", i+1, err)
		}

		// // Validate volume mounts and volumes for common pitfalls (subPath leading slash, empty PVC claimName)
		// if err := validateVolumeMounts(obj); err != nil {
		// 	return nil, fmt.Errorf("validation failed in document %d: %w", i+1, err)
		// }

		// Validate K8s Spec (Structure)
		gvk, name, err := k8s.ValidateK8sJSON(jsonBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to validate K8s spec for document %d: %w", i+1, err)
		}

		resourcesToCreate = append(resourcesToCreate, &resource.Resource{
			Type:       resource.ResourceType(normalizeResourceKind(gvk.Kind)),
			Name:       name,
			ParsedYAML: datatypes.JSON(jsonBytes),
		})
	}
	return resourcesToCreate, nil
}

// validateVolumeMounts checks PodSpecs for common volume/volumeMount mistakes.
// - disallow `subPath` that starts with '/'
// - ensure persistentVolumeClaim.claimName (if present) is non-empty and not absolute
// func validateVolumeMounts(obj map[string]interface{}) error {
// 	podSpecs := findPodSpecs(obj)
// 	for _, spec := range podSpecs {
// 		// collect volumes by name for lookup
// 		volumesMap := map[string]map[string]interface{}{}
// 		if vols, ok := spec["volumes"].([]interface{}); ok {
// 			for _, v := range vols {
// 				if vm, ok := v.(map[string]interface{}); ok {
// 					if name, ok := vm["name"].(string); ok && name != "" {
// 						volumesMap[name] = vm
// 					}
// 				}
// 			}
// 		}

// 		// check each container's mounts
// 		containers := getContainersFromPodSpec(spec)
// 		for _, c := range containers {
// 			if vms, ok := c["volumeMounts"].([]interface{}); ok {
// 				for _, vm := range vms {
// 					if vmMap, ok := vm.(map[string]interface{}); ok {
// 						if subPathRaw, ok := vmMap["subPath"].(string); ok {
// 							if strings.HasPrefix(subPathRaw, "/") {
// 								return fmt.Errorf("%w: volumeMount.subPath must not start with '/': %s", ErrInvalidVolumeMounts, subPathRaw)
// 							}
// 						}
// 						// if mount references a volume, check the backing volume for PVC claimName
// 						if volName, ok := vmMap["name"].(string); ok && volName != "" {
// 							if volDef, found := volumesMap[volName]; found {
// 								if pvc, ok := volDef["persistentVolumeClaim"].(map[string]interface{}); ok {
// 									if claimNameRaw, ok := pvc["claimName"].(string); ok {
// 										if claimNameRaw == "" {
// 											return fmt.Errorf("%w: persistentVolumeClaim.claimName for volume '%s' is empty", ErrInvalidVolumeMounts, volName)
// 										}
// 										if strings.HasPrefix(claimNameRaw, "/") {
// 											return fmt.Errorf("%w: persistentVolumeClaim.claimName must not start with '/': %s", ErrInvalidVolumeMounts, claimNameRaw)
// 										}
// 										if strings.Contains(claimNameRaw, "{{") || strings.Contains(claimNameRaw, "}}") {
// 											return fmt.Errorf("%w: persistentVolumeClaim.claimName appears templated or invalid: %s", ErrInvalidVolumeMounts, claimNameRaw)
// 										}
// 									} else {
// 										return fmt.Errorf("persistentVolumeClaim.claimName for volume '%s' is not a string", volName)
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func validateContainerLimits(obj map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	podSpecs := findPodSpecs(obj)
	for _, podSpec := range podSpecs {
		containers := getContainersFromPodSpec(podSpec)
		for _, container := range containers {
			if err := checkSingleContainerLimits(container); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSingleContainerLimits(container map[string]interface{}) error {
	containerName, _ := container["name"].(string)
	resources, ok := container["resources"].(map[string]interface{})
	if !ok {
		return nil // No resources defined is strictly valid K8s, though maybe not best practice
	}

	requests, _ := resources["requests"].(map[string]interface{})
	limits, _ := resources["limits"].(map[string]interface{})

	if requests == nil || limits == nil {
		return nil
	}

	checkResource := func(resName string) error {
		reqStr, hasReq := getStringValue(requests, resName)
		limStr, hasLim := getStringValue(limits, resName)

		if hasReq && hasLim {
			reqQ, err1 := k8sRes.ParseQuantity(reqStr)
			limQ, err2 := k8sRes.ParseQuantity(limStr)

			if err1 == nil && err2 == nil {
				if limQ.Cmp(reqQ) < 0 {
					return fmt.Errorf("container '%s': %s limit (%s) cannot be less than request (%s)",
						containerName, resName, limStr, reqStr)
				}
			}
		}
		return nil
	}

	if err := checkResource("cpu"); err != nil {
		return err
	}
	if err := checkResource("memory"); err != nil {
		return err
	}
	return nil
}

func normalizeResourceKind(kind string) string {
	switch strings.ToLower(kind) {
	case "pod":
		return "Pod"
	case "service":
		return "Service"
	case "deployment":
		return "Deployment"
	case "configmap":
		return "ConfigMap"
	case "ingress":
		return "Ingress"
	case "job":
		return "Job"
	case "cronjob":
		return "CronJob"
	default:
		if kind == "" {
			return kind
		}
		// Capitalize first letter as fallback
		return strings.ToUpper(string(kind[0])) + strings.ToLower(kind[1:])
	}
}
