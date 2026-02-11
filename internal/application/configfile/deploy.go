package configfile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/datatypes"
)

// CreateInstance deploys resources to Kubernetes with a high-performance pipeline.
func (s *ConfigFileService) CreateInstance(ctx context.Context, id string, claims *types.Claims) error {
	// 1. Fetch Data
	resources, err := s.Repos.Resource.ListResourcesByConfigFileID(id)
	if err != nil {
		return err
	}
	cf, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return err
	}

	// 2. Prepare Context (Namespace, Project, Claims)
	ns, proj, err := s.prepareNamespaceAndProject(ctx, cf, claims)
	if err != nil {
		return err
	}

	// 3. Determine Deployment Strategy (Job-only vs Standard)
	// isJobOnly := s.configFileIsAllJobs(resources)

	// 4. Prepare Variables & Volumes
	var templateValues map[string]string
	var shouldEnforceRO bool
	var groupPVCName string

	// Standard Deployment: Bind Volumes & Check Permissions
	userPvc, grpPvc := s.bindProjectAndUserVolumes(ns, proj, claims)
	shouldEnforceRO, err = s.determineReadOnlyEnforcement(claims, proj)
	if err != nil {
		return err
	}
	groupPVCName = grpPvc
	templateValues = s.buildTemplateValues(cf, ns, userPvc, grpPvc, claims)

	// 5. Processing Pipeline (The most compute-intensive part)
	// We use pre-allocation to avoid slice resizing overhead
	type processedResource struct {
		Name     string
		Kind     string
		JSONData []byte
	}
	processedResources := make([]processedResource, 0, len(resources))

	for _, res := range resources {
		// A. Template Replacement (String Level)
		jsonStr := string(res.ParsedYAML)
		replacedJSON, err := utils.ReplacePlaceholdersInJSON(jsonStr, templateValues)
		if err != nil {
			return fmt.Errorf("failed to replace placeholders for resource %s: %w", res.Name, err)
		}

		// B. Unmarshal ONCE (Performance Key)
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(replacedJSON), &obj); err != nil {
			return fmt.Errorf("failed to unmarshal resource %s: %w", res.Name, err)
		}

		// C. Apply Patches (In-Memory Map Manipulation)
		//    All business logic validation and injection happens here without re-marshaling.
		patchCtx := &PatchContext{
			ProjectID:       cf.ProjectID,
			Project:         proj,
			UserIsAdmin:     claims.IsAdmin,
			ShouldEnforceRO: shouldEnforceRO,
			GroupPVC:        groupPVCName,
		}

		if err := s.applyResourcePatches(obj, patchCtx); err != nil {
			return fmt.Errorf("failed to patch resource %s: %w", res.Name, err)
		}

		// Extract kind and name BEFORE marshaling (avoid double unmarshal)
		kind, _ := obj["kind"].(string)
		metadata, _ := obj["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)

		// D. Marshal ONCE
		finalBytes, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("failed to marshal final resource %s: %w", res.Name, err)
		}

		processedResources = append(processedResources, processedResource{
			Name:     name,
			Kind:     kind,
			JSONData: finalBytes,
		})
	}

	// 6. Apply to Kubernetes
	slog.Info("deploying resources to namespace",
		"resource_count", len(processedResources),
		"namespace", ns)

	// Use executor if available, otherwise use direct K8s deployment
	if s.executor != nil {
		// Generate job ID
		jobID, err := gonanoid.New()
		if err != nil {
			return fmt.Errorf("failed to generate job ID: %w", err)
		}

		// Prepare resource payloads (no double unmarshal needed!)
		payloads := make([]executor.ResourcePayload, len(processedResources))
		for i, pr := range processedResources {
			payloads[i] = executor.ResourcePayload{
				Name:     pr.Name,
				Kind:     pr.Kind,
				JSONData: pr.JSONData,
			}
		}

		// Submit job via executor
		_, err = s.executor.Submit(ctx, &executor.SubmitRequest{
			JobID:        jobID,
			ConfigFileID: id,
			ProjectID:    cf.ProjectID,
			Namespace:    ns,
			UserID:       claims.UserID,
			Username:     claims.Username,
			Resources:    payloads,
		})
		if err != nil {
			return fmt.Errorf("failed to submit job: %w", err)
		}
	} else {
		// Fallback to direct K8s deployment
		for _, pr := range processedResources {
			if err := k8s.CreateByJson(datatypes.JSON(pr.JSONData), ns); err != nil {
				return fmt.Errorf("failed to create resource in k8s: %w", err)
			}
		}
	}

	return nil
}

func (s *ConfigFileService) DeleteInstance(ctx context.Context, id string, claims *types.Claims) error {
	data, err := s.Repos.Resource.ListResourcesByConfigFileID(id)
	if err != nil {
		return err
	}
	configfile, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return err
	}

	safeUsername := k8s.ToSafeK8sName(claims.Username)
	ns := k8s.FormatNamespaceName(configfile.ProjectID, safeUsername)

	for _, val := range data {
		if err := k8s.DeleteByJson(val.ParsedYAML, ns); err != nil {
			// Continue deleting other resources even if one fails
			slog.Error("failed to delete resource",
				"resource", val.Name,
				"error", err)
		}
	}
	return nil
}

func (s *ConfigFileService) DeleteConfigFileInstance(id string) error {
	configfile, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return err
	}

	resources, err := s.Repos.Resource.ListResourcesByConfigFileID(id)
	if err != nil {
		return err
	}

	users, err := s.Repos.User.ListUsersByProjectID(configfile.ProjectID)
	if err != nil {
		return err
	}

	for _, user := range users {
		safeUsername := k8s.ToSafeK8sName(user.Username)
		ns := k8s.FormatNamespaceName(configfile.ProjectID, safeUsername)
		for _, res := range resources {
			if err := k8s.DeleteByJson(res.ParsedYAML, ns); err != nil {
				slog.Warn("failed to delete instance for user",
					"username", user.Username,
					"namespace", ns,
					"error", err)
			}
		}
	}

	return nil
}
