package configfile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/datatypes"
)

// CreateInstance deploys resources to Kubernetes with a high-performance pipeline.
func (s *ConfigFileService) CreateInstance(ctx context.Context, id string, claims *types.Claims) error {
	// 1. Fetch Data
	resources, err := s.Repos.Resource.ListResourcesByCommitID(ctx, id)
	if err != nil {
		return err
	}
	commit, err := s.Repos.ConfigFile.GetCommit(ctx, id)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		blob, err := s.Repos.ConfigFile.GetBlob(ctx, commit.BlobHash)
		if err != nil {
			return err
		}
		var rawYaml string
		if err := json.Unmarshal(blob.Content, &rawYaml); err != nil {
			rawYaml = string(blob.Content)
		}
		parsedResources, err := s.parseAndValidateResources(rawYaml)
		if err != nil {
			return err
		}
		resources = make([]resource.Resource, len(parsedResources))
		for i, res := range parsedResources {
			if res != nil {
				resources[i] = *res
			}
		}
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
	ns, proj, err := s.prepareNamespaceAndProject(ctx, commit.ProjectID, claims)
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
			jobID = uuid.NewString()
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
	templateValues = s.buildTemplateValues(commit.ProjectID, ns, userPvc, grpPvc, claims)

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
			injectJobLabels(obj, jobID, commit.ID)
		}

		// C. Apply Patches (In-Memory Map Manipulation)
		//    All business logic validation and injection happens here without re-marshaling.
		patchCtx := &PatchContext{
			ProjectID:       commit.ProjectID,
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
			JobID:          jobID,
			ConfigCommitID: id,
			ProjectID:      commit.ProjectID,
			Namespace:      ns,
			UserID:         claims.UserID,
			Username:       claims.Username,
			Resources:      payloads,
			SubmitType:     submitTypeValue,
			QueueName:      queueName,
			Priority:       priority,
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
