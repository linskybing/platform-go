package image

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	cfg "github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/image"
)

func (s *ImageService) ApproveRequest(id uint, note string, isGlobal bool, approverID uint) error {
	req, err := s.repo.FindRequestByID(id)
	if err != nil {
		return err
	}

	// If approver marked this as global, clear ProjectID so the created
	// allow-list rule will be global (ProjectID == nil)
	if isGlobal {
		req.ProjectID = nil
	}

	req.Status = "approved"
	req.ReviewerNote = note
	req.ReviewerID = &approverID
	req.ReviewedAt = ptrTime(time.Now())

	if err := s.repo.UpdateRequest(req); err != nil {
		return err
	}

	return s.createCoreAndPolicyFromRequest(req, approverID)
}

func (s *ImageService) createCoreAndPolicyFromRequest(req *image.ImageRequest, adminID uint) error {
	fullName := req.InputImageName
	if req.InputRegistry != "" && req.InputRegistry != "docker.io" {
		fullName = fmt.Sprintf("%s/%s", req.InputRegistry, req.InputImageName)
	}

	parts := strings.Split(req.InputImageName, "/")
	var namespace, name string
	if len(parts) >= 2 {
		namespace = parts[0]
		name = strings.Join(parts[1:], "/")
	} else {
		namespace = "library"
		name = req.InputImageName
	}

	repoEntity := &image.ContainerRepository{
		Registry:  req.InputRegistry,
		Namespace: namespace,
		Name:      name,
		FullName:  fullName,
	}
	if err := s.repo.FindOrCreateRepository(repoEntity); err != nil {
		return err
	}

	tagEntity := &image.ContainerTag{
		RepositoryID: repoEntity.ID,
		Name:         req.InputTag,
	}
	if err := s.repo.FindOrCreateTag(tagEntity); err != nil {
		return err
	}

	rule := &image.ImageAllowList{
		ProjectID:    req.ProjectID,
		RepositoryID: repoEntity.ID,
		TagID:        &tagEntity.ID,
		RequestID:    &req.ID,
		CreatedBy:    adminID,
		IsEnabled:    true,
	}

	if err := s.repo.CreateAllowListRule(rule); err != nil {
		return err
	}

	// If the image is already present in our Harbor private registry, mark it as pulled
	// so that injection logic (injectHarborPrefix) and UI reflect the image as available.
	fullNameLower := strings.ToLower(fullName)
	harborPrefixLower := strings.ToLower(cfg.HarborPrivatePrefix)
	if strings.HasPrefix(fullNameLower, harborPrefixLower) {
		status := &image.ClusterImageStatus{
			TagID:        tagEntity.ID,
			IsPulled:     true,
			LastPulledAt: ptrTime(time.Now()),
		}
		if err := s.repo.UpdateClusterStatus(status); err != nil {
			slog.Error("failed to mark image as pulled",
				"image", fullName,
				"tag", req.InputTag,
				"error", err)
		}
	}

	return nil
}
