package configfile

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/project"
)

type PatchContext struct {
	ProjectID       string
	Project         project.Project
	UserIsAdmin     bool
	ShouldEnforceRO bool
	GroupPVC        string
}

// applyResourcePatches orchestrates all modifications to the K8s object map.
func (s *ConfigFileService) applyResourcePatches(obj map[string]interface{}, ctx *PatchContext) error {
	// 0. Apply job-specific patches before scanning for pod specs.
	if err := s.patchJobDeadline(obj, ctx); err != nil {
		return err
	}

	// 1. Identify Pod Specs once to avoid traversing the tree multiple times
	podSpecs := findPodSpecs(obj)

	if len(podSpecs) == 0 {
		return nil // Non-workload resources (Services, ConfigMaps) usually don't need patches
	}

	// 2. Iterate PodSpecs and apply patches
	for _, spec := range podSpecs {
		// A. Validate & Patch Images
		if err := s.patchImages(spec, ctx); err != nil {
			return err
		}

		// B. Enforce ReadOnly PVCs
		if ctx.ShouldEnforceRO && ctx.GroupPVC != "" {
			s.patchReadOnly(spec, ctx.GroupPVC)
		}

		// C. Inject GPU Config
		if err := s.patchGPU(spec, ctx.Project); err != nil {
			return err
		}

		// D. Inject General Security Context
		s.patchSecurityContext(spec)

		// E. Inject Scheduler Priority and Queue
		s.patchPriority(spec)
	}

	return nil
}

func (s *ConfigFileService) patchJobDeadline(obj map[string]interface{}, ctx *PatchContext) error {
	if obj == nil || ctx == nil {
		return nil
	}
	if ctx.Project.MaxJobRuntimeSeconds <= 0 {
		return nil
	}
	kind, _ := obj["kind"].(string)
	deadline := int64(ctx.Project.MaxJobRuntimeSeconds)
	if kind == "Job" {
		spec, ok := obj["spec"].(map[string]interface{})
		if !ok {
			return nil
		}
		spec["activeDeadlineSeconds"] = deadline
		return nil
	}
	if kind == "CronJob" {
		spec, ok := obj["spec"].(map[string]interface{})
		if !ok {
			return nil
		}
		jobTemplate, ok := spec["jobTemplate"].(map[string]interface{})
		if !ok {
			return nil
		}
		jobSpec, ok := jobTemplate["spec"].(map[string]interface{})
		if !ok {
			return nil
		}
		jobSpec["activeDeadlineSeconds"] = deadline
	}
	return nil
}

func (s *ConfigFileService) patchImages(podSpec map[string]interface{}, ctx *PatchContext) error {
	containers := getContainersFromPodSpec(podSpec)

	for _, cont := range containers {
		img, ok := cont["image"].(string)
		if !ok || img == "" {
			continue
		}

		// 1. Validation (Skip if Admin)
		if !ctx.UserIsAdmin && s.imageSvc != nil {
			imageName, imageTag := parseImageNameTag(img)
			allowed, err := s.imageSvc.ValidateImageForProject(imageName, imageTag, &ctx.ProjectID)
			slog.Debug("validating image for project",
				"image", imageName,
				"tag", imageTag,
				"allowed", allowed,
				"error", err)
			if err != nil {
				return fmt.Errorf("failed to validate image %s: %v", img, err)
			}
			if !allowed {
				return fmt.Errorf("image '%s:%s' is not allowed for this project", imageName, imageTag)
			}
		}

		// 2. Harbor Injection
		// Logic: Check if image needs pulling from private Harbor
		// Note: parseImageNameTag is lightweight, safe to recall
		imageName, imageTag := parseImageNameTag(img)

		// Optimization: Check prefix first
		if strings.HasPrefix(img, config.HarborPrivatePrefix) {
			continue
		}

		if s.imageSvc == nil {
			continue
		}
		allowedImg, err := s.imageSvc.GetAllowedImage(imageName, imageTag, ctx.ProjectID)
		if err == nil && allowedImg != nil && allowedImg.IsPulled {
			cont["image"] = config.HarborPrivatePrefix + img
		}
	}
	return nil
}

func (s *ConfigFileService) patchReadOnly(podSpec map[string]interface{}, targetPvcName string) {
	// Identify volumes pointing to the restricted PVC
	targetVolumes := make(map[string]bool)
	if volumes, ok := podSpec["volumes"].([]interface{}); ok {
		for _, v := range volumes {
			if vol, ok := v.(map[string]interface{}); ok {
				if pvcSource, ok := vol["persistentVolumeClaim"].(map[string]interface{}); ok {
					if claimName, _ := pvcSource["claimName"].(string); claimName == targetPvcName {
						volName, _ := vol["name"].(string)
						targetVolumes[volName] = true
					}
				}
			}
		}
	}

	if len(targetVolumes) == 0 {
		return
	}

	// Enforce ReadOnly on matched volume mounts
	containers := getContainersFromPodSpec(podSpec)
	for _, c := range containers {
		if mounts, ok := c["volumeMounts"].([]interface{}); ok {
			for _, m := range mounts {
				if mount, ok := m.(map[string]interface{}); ok {
					if mountName, _ := mount["name"].(string); targetVolumes[mountName] {
						mount["readOnly"] = true
					}
				}
			}
		}
	}
}

func (s *ConfigFileService) patchGPU(podSpec map[string]interface{}, p project.Project) error {
	// 1. Check if GPU is requested
	hasGPU := false
	containers := getContainersFromPodSpec(podSpec)

	// Fast check loop
	for _, c := range containers {
		if resources, ok := c["resources"].(map[string]interface{}); ok {
			if requests, ok := resources["requests"].(map[string]interface{}); ok {
				if _, exists := requests["nvidia.com/gpu"]; exists {
					hasGPU = true
					break
				}
			}
		}
	}

	if !hasGPU {
		return nil
	}

	// 2. Validate Project Quota
	if p.GPUQuota <= 0 {
		return fmt.Errorf("project GPU configuration invalid: GPUQuota=%d. Must be greater than 0", p.GPUQuota)
	}

	// 3. Inject Config
	for _, c := range containers {
		resources, ok := c["resources"].(map[string]interface{})
		if !ok {
			continue
		}
		requests, ok := resources["requests"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check if GPU is requested
		reqVal, exists := requests["nvidia.com/gpu"]
		if !exists {
			continue
		}

		// Parse requested GPU count (handle string or int/float input safely)
		requestedGPU, err := strconv.Atoi(fmt.Sprintf("%v", reqVal))
		if err != nil {
			// Handle error or skip if value is invalid
			continue
		}

		// Calculate the final GPU count: min(requested, quota)
		// This logic ensures that if request=1 and quota=20, final is 1 (fixing the K8s error).
		// If request=50 and quota=20, final is 20 (enforcing quota).
		finalGPU := requestedGPU
		quota := int(p.GPUQuota)

		if quota < finalGPU {
			finalGPU = quota
		}

		// Kubernetes requires nvidia.com/gpu requests and limits to be equal
		gpuQtyStr := fmt.Sprintf("%d", finalGPU)

		// Update requests
		requests["nvidia.com/gpu"] = gpuQtyStr

		// Update limits
		limits, ok := resources["limits"].(map[string]interface{})
		if !ok {
			limits = make(map[string]interface{})
			resources["limits"] = limits
		}
		limits["nvidia.com/gpu"] = gpuQtyStr
	}
	return nil
}

func (s *ConfigFileService) patchSecurityContext(podSpec map[string]interface{}) {
	const targetUID int64 = 0
	const targetGID int64 = 0

	secContext, ok := podSpec["securityContext"].(map[string]interface{})
	if !ok {
		secContext = make(map[string]interface{})
		podSpec["securityContext"] = secContext
	}

	secContext["runAsUser"] = targetUID
	secContext["runAsGroup"] = targetGID

	// Inject fsGroup if volumes exist
	if _, hasVolumes := podSpec["volumes"]; hasVolumes {
		secContext["fsGroup"] = targetUID
		secContext["fsGroupChangePolicy"] = "OnRootMismatch"
	}
}

func (s *ConfigFileService) patchPriority(podSpec map[string]interface{}) {
	if config.ConfigFilePriorityClassName != "" {
		podSpec["priorityClassName"] = config.ConfigFilePriorityClassName
	}
	if !config.FlashSchedEnabled && config.ExecutorMode != "scheduler" {
		return
	}

	if config.SchedulerName != "" {
		podSpec["schedulerName"] = config.SchedulerName
	}

	annotations := ensurePodAnnotations(podSpec)
	if config.FlashSchedQueueAnnotationKey != "" && config.ConfigFileQueueName != "" {
		annotations[config.FlashSchedQueueAnnotationKey] = config.ConfigFileQueueName
	}
	if config.FlashSchedPreemptableAnnotationKey != "" {
		annotations[config.FlashSchedPreemptableAnnotationKey] = strconv.FormatBool(config.ConfigFilePreemptable)
	}
}

func ensurePodAnnotations(podSpec map[string]interface{}) map[string]interface{} {
	if annotations, ok := podSpec["annotations"].(map[string]interface{}); ok {
		return annotations
	}
	if metadata, ok := podSpec["metadata"].(map[string]interface{}); ok {
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			return annotations
		}
		annotations := make(map[string]interface{})
		metadata["annotations"] = annotations
		return annotations
	}
	annotations := make(map[string]interface{})
	podSpec["annotations"] = annotations
	return annotations
}
