package image

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	cfg "github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/internal/repository"
)

type ImageService struct {
	repo repository.ImageRepo
}

func NewImageService(repo repository.ImageRepo) *ImageService {
	return &ImageService{repo: repo}
}

func (s *ImageService) SubmitRequest(userID string, registry, name, tag string, projectID *string) (*image.ImageRequest, error) {
	// If caller didn't provide a registry, try to parse it out from the name
	// (e.g. "192.168.110.1:30003/library/pros-cameraapi" -> registry: "192.168.110.1:30003", name: "library/pros-cameraapi").
	// If no registry is found, leave InputRegistry empty to represent Docker Hub/default behavior.
	parsedRegistry := registry
	parsedName := name
	if parsedRegistry == "" {
		parts := strings.SplitN(name, "/", 2)
		first := parts[0]
		hasDomain := strings.Contains(first, ".") || strings.Contains(first, ":") || first == "localhost"
		if hasDomain && len(parts) == 2 {
			parsedRegistry = first
			parsedName = parts[1]
		}
	}

	req := &image.ImageRequest{
		UserID:         userID,
		ProjectID:      projectID,
		InputRegistry:  parsedRegistry,
		InputImageName: parsedName,
		InputTag:       tag,
		Status:         "pending",
	}

	if warn := s.validateNameAndTag(name, tag); warn != "" {
		slog.Warn("image validation warning",
			"warning", warn,
			"image", name,
			"tag", tag)
	}

	// Build fullname for allow-list checks
	fullName := parsedName
	if parsedRegistry != "" && parsedRegistry != "docker.io" {
		fullName = fmt.Sprintf("%s/%s", parsedRegistry, parsedName)
	}

	// If an enabled allow-list rule already exists (global or project-scoped),
	// do not create a duplicate request.
	allowed, err := s.repo.CheckImageAllowed(projectID, fullName, tag)
	if err != nil {
		return req, err
	}
	if allowed {
		return nil, fmt.Errorf("image '%s:%s' is already allowed for this project", fullName, tag)
	}

	if err := s.repo.CreateRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

func (s *ImageService) ListRequests(projectID *string, status string) ([]image.ImageRequest, error) {
	return s.repo.ListRequests(projectID, status)
}

func (s *ImageService) RejectRequest(id string, note string, approverID string) (*image.ImageRequest, error) {
	req, err := s.repo.FindRequestByID(id)
	if err != nil {
		return nil, err
	}
	req.Status = "rejected"
	req.ReviewerNote = note
	req.ReviewerID = &approverID
	req.ReviewedAt = ptrTime(time.Now())

	if err := s.repo.UpdateRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ImageService) GetAllowedImage(name, tag string, projectID string) (*image.AllowedImageDTO, error) {
	rule, err := s.repo.FindAllowListRule(&projectID, name, tag)
	if err != nil {
		return nil, err
	}

	status, _ := s.repo.GetClusterStatus(rule.Tag.ID)
	isPulled := false
	if status != nil {
		isPulled = status.IsPulled
	}

	return &image.AllowedImageDTO{
		ID:        rule.ID,
		Registry:  rule.Repository.Registry,
		ImageName: rule.Repository.Name,
		Tag:       rule.Tag.Name,
		Digest:    rule.Tag.Digest,
		ProjectID: rule.ProjectID,
		IsGlobal:  rule.ProjectID == nil,
		IsPulled:  isPulled,
	}, nil
}

func (s *ImageService) ListAllowedImages(projectID *string) ([]image.AllowedImageDTO, error) {
	rules, err := s.repo.ListAllowedImages(projectID)
	if err != nil {
		return nil, err
	}

	var dtos []image.AllowedImageDTO
	for _, rule := range rules {
		isGlobal := rule.ProjectID == nil

		status, _ := s.repo.GetClusterStatus(rule.Tag.ID)
		isPulled := false
		if status != nil {
			isPulled = status.IsPulled
		}

		displayImageName := rule.Repository.FullName
		if displayImageName == "" {
			if rule.Repository.Namespace != "" {
				displayImageName = fmt.Sprintf("%s/%s", rule.Repository.Namespace, rule.Repository.Name)
			} else {
				displayImageName = rule.Repository.Name
			}
		}

		dtos = append(dtos, image.AllowedImageDTO{
			ID:        rule.ID,
			Registry:  rule.Repository.Registry,
			ImageName: displayImageName,
			Tag:       rule.Tag.Name,
			Digest:    rule.Tag.Digest,
			ProjectID: rule.ProjectID,
			IsGlobal:  isGlobal,
			IsPulled:  isPulled,
		})
	}
	return dtos, nil
}

func (s *ImageService) AddProjectImage(userID string, projectID string, name, tag string) error {
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		return fmt.Errorf("invalid image format: %s", warn)
	}

	req := &image.ImageRequest{
		UserID:         userID,
		ProjectID:      &projectID,
		InputImageName: name,
		InputTag:       tag,
		Status:         "pending", // Needs approval
	}

	return s.repo.CreateRequest(req)
}

func (s *ImageService) ValidateImageForProject(name, tag string, projectID *string) (bool, error) {
	fullName := name
	// If the image already points to our Harbor private registry and the request
	// is global (projectID == nil), consider it allowed automatically.
	// For project-scoped requests (projectID != nil), do NOT auto-allow — require admin approval.
	if cfg.HarborPrivatePrefix != "" {
		lowerFull := strings.ToLower(fullName)
		lowerPrefix := strings.ToLower(cfg.HarborPrivatePrefix)
		if strings.HasPrefix(lowerFull, lowerPrefix) && projectID == nil {
			return true, nil
		}
	}

	return s.repo.CheckImageAllowed(projectID, fullName, tag)
}

func (s *ImageService) GetPullJobStatus(jobID string) *PullJobStatus {
	return pullTracker.GetJob(jobID)
}

func (s *ImageService) SubscribeToPullJob(jobID string) <-chan *PullJobStatus {
	return pullTracker.Subscribe(jobID)
}

func (s *ImageService) GetFailedPullJobs(limit int) []*PullJobStatus {
	return pullTracker.GetFailedJobs(limit)
}

func (s *ImageService) GetActivePullJobs() []*PullJobStatus {
	return pullTracker.GetActiveJobs()
}

func (s *ImageService) validateNameAndTag(name, tag string) string {
	name = strings.TrimSpace(name)
	tag = strings.TrimSpace(tag)
	if name == "" || tag == "" {
		return "image name/tag should not be empty"
	}

	nameRe := regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*$`)
	if !nameRe.MatchString(name) {
		return "image name format looks invalid"
	}

	tagRe := regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
	if !tagRe.MatchString(tag) {
		return "image tag format looks invalid"
	}

	return ""
}

func (s *ImageService) DisableAllowListRule(id string) error {
	return s.repo.DisableAllowListRule(id)
}
