package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/domain/configfile"
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

// syncConfigFileResources manages the diff (create/update/delete) for config file updates.
func (s *ConfigFileService) syncConfigFileResources(c *gin.Context, cf *configfile.ConfigFile, rawYaml string, newResources []*resource.Resource) error {
	// 1. Fetch existing resources
	existingResources, err := s.Repos.Resource.ListResourcesByConfigFileID(cf.CFID)
	if err != nil {
		return err
	}

	existingMap := make(map[string]resource.Resource)
	for _, r := range existingResources {
		existingMap[r.Name] = r
	}

	// 2. Track which existing resources are still present
	processedNames := make(map[string]bool)

	// 3. Update or Create
	for i, newRes := range newResources {
		name := newRes.Name
		processedNames[name] = true

		if val, exists := existingMap[name]; exists {
			// Update
			oldTarget := val
			val.Name = name
			val.ParsedYAML = newRes.ParsedYAML
			val.Type = newRes.Type // Ensure type is updated if kind changed (rare but possible)

			fmt.Printf("Updating resource for document %d: %s\n", i+1, name)
			if err := s.Repos.Resource.UpdateResource(&val); err != nil {
				return fmt.Errorf("failed to update resource %s: %w", name, err)
			}
			utils.LogAuditWithConsole(c, "update", "resource", fmt.Sprintf("r_id=%d", val.RID), oldTarget, val, "", s.Repos.Audit)
		} else {
			// Create
			newRes.CFID = cf.CFID
			fmt.Printf("Creating resource for document %d: %s\n", i+1, name)
			if err := s.Repos.Resource.CreateResource(newRes); err != nil {
				return fmt.Errorf("failed to create resource %s: %w", name, err)
			}
			utils.LogAuditWithConsole(c, "create", "resource", fmt.Sprintf("r_id=%d", newRes.RID), nil, *newRes, "", s.Repos.Audit)
		}
	}

	// 4. Delete removed resources
	for name, res := range existingMap {
		if !processedNames[name] {
			// Remove from DB (Instance cleanup happens separately via re-deploy usually, or should be handled here if strictly synced)
			// Note: This logic assumes the instance is cleaned up via deleteConfigFileInstance call in Service before this,
			// or will be updated by next Apply.
			if err := s.Repos.Resource.DeleteResource(res.RID); err != nil {
				return fmt.Errorf("failed to delete unused resource %s: %w", name, err)
			}
			utils.LogAuditWithConsole(c, "delete", "resource", fmt.Sprintf("r_id=%d", res.RID), res, nil, "", s.Repos.Audit)
		}
	}
	return nil
}

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
