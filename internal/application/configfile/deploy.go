package configfile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
	gonanoid "github.com/matoous/go-nanoid/v2"
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

	submitType := string(executor.SubmitTypeJob)
	if ctxSubmitType, ok := submitTypeFromContext(ctx); ok {
		submitType = ctxSubmitType
	}
	resources, err = filterResourcesBySubmitType(resources, submitType)
	if err != nil {
		return err
	}

	// 2. Prepare Context (Namespace, Project, Claims)
	ns, proj, err := s.prepareNamespaceAndProject(ctx, cf, claims)
	if err != nil {
		return err
	}
	if allowed, err := proj.IsTimeAllowed(time.Now()); err != nil {
		return fmt.Errorf("failed to validate project schedule: %w", err)
	} else if !allowed {
		return fmt.Errorf("project is outside allowed schedule")
	}

	// 3. Determine Deployment Strategy (Job-only vs Standard)
	// isJobOnly := s.configFileIsAllJobs(resources)

	var jobID string
	var queueName string
	var priority int32
	if s.executor != nil {
		if ctxJobID, ok := jobIDFromContext(ctx); ok {
			jobID = ctxJobID
		} else {
			jobIDValue, err := gonanoid.New()
			if err != nil {
				return fmt.Errorf("failed to generate job ID: %w", err)
			}
			jobID = jobIDValue
		}
		if ctxQueueName, ok := queueNameFromContext(ctx); ok {
			queueName = ctxQueueName
		}
		if ctxPriority, ok := priorityFromContext(ctx); ok {
			priority = ctxPriority
		}
		if queueName == "" {
			queueName = defaultQueueForResources(resources)
		}
	}

	// 4. Prepare Variables & Volumes
	var templateValues map[string]string
	var shouldEnforceRO bool
	var groupPVCName string

	// Standard Deployment: Bind Volumes & Check Permissions
	userPvc, grpPvc, err := s.bindProjectAndUserVolumes(ctx, ns, proj, claims, resources)
	if err != nil {
		return err
	}
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

		if jobID != "" {
			injectJobLabels(obj, jobID, cf.CFID)
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
		submitTypeValue := executor.SubmitType(submitType)
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
			SubmitType:   submitTypeValue,
			QueueName:    queueName,
			Priority:     priority,
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

func filterResourcesBySubmitType(resources []resource.Resource, submitType string) ([]resource.Resource, error) {
	if submitType == "" {
		return resources, nil
	}
	mode := strings.ToLower(submitType)
	if mode != string(executor.SubmitTypeJob) && mode != string(executor.SubmitTypeWorkflow) {
		return nil, fmt.Errorf("invalid submit_type: %s", submitType)
	}

	filtered := make([]resource.Resource, 0, len(resources))
	workloadCount := 0
	for _, res := range resources {
		kind := strings.ToLower(string(res.Type))
		if kind == "" {
			filtered = append(filtered, res)
			continue
		}
		if isJobWorkloadKind(kind) {
			if mode == string(executor.SubmitTypeJob) {
				filtered = append(filtered, res)
				workloadCount++
			}
			continue
		}
		if isWorkflowWorkloadKind(kind) {
			if mode == string(executor.SubmitTypeWorkflow) {
				filtered = append(filtered, res)
				workloadCount++
			}
			continue
		}
		filtered = append(filtered, res)
	}
	if workloadCount == 0 {
		return nil, fmt.Errorf("no %s workload resources found in configfile", mode)
	}
	return filtered, nil
}

func isJobWorkloadKind(kind string) bool {
	switch kind {
	case "job", "cronjob", "flashjob":
		return true
	default:
		return false
	}
}

func isWorkflowWorkloadKind(kind string) bool {
	switch kind {
	case "workflow", "workflowtemplate", "cronworkflow":
		return true
	default:
		return false
	}
}
