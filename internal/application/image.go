package application

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImageService struct {
	repo repository.ImageRepo
}

func NewImageService(repo repository.ImageRepo) *ImageService {
	return &ImageService{repo: repo}
}

func (s *ImageService) SubmitRequest(userID uint, name, tag string, projectID *uint) (*image.ImageRequest, error) {
	req := &image.ImageRequest{
		UserID:    userID,
		Name:      name,
		Tag:       tag,
		ProjectID: projectID,
		Status:    "pending",
	}
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		log.Printf("[image-validate] warning: %s", warn)
		req.Note = warn
	}
	return req, s.repo.CreateRequest(req)
}

func (s *ImageService) ListRequests(status string) ([]image.ImageRequest, error) {
	return s.repo.ListRequests(status)
}

func (s *ImageService) ApproveRequest(id uint, note string, isGlobal bool) (*image.ImageRequest, error) {
	req, err := s.repo.FindRequestByID(id)
	if err != nil {
		return nil, err
	}
	if warn := s.validateNameAndTag(req.Name, req.Tag); warn != "" && req.Note == "" {
		log.Printf("[image-validate] warning on approve: %s", warn)
		req.Note = warn
	}
	req.Status = "approved"
	req.Note = note
	if err := s.repo.UpdateRequest(req); err != nil {
		return nil, err
	}

	// Add to allowed images
	allowedImg := &image.AllowedImage{
		Name:      req.Name,
		Tag:       req.Tag,
		ProjectID: req.ProjectID,
		IsGlobal:  isGlobal,
		CreatedBy: req.UserID,
	}
	_ = s.repo.CreateAllowed(allowedImg)
	return req, nil
}

func (s *ImageService) RejectRequest(id uint, note string) (*image.ImageRequest, error) {
	req, err := s.repo.FindRequestByID(id)
	if err != nil {
		return nil, err
	}
	req.Status = "rejected"
	req.Note = note
	if err := s.repo.UpdateRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ImageService) ListAllowed() ([]image.AllowedImage, error) {
	return s.repo.ListAllowed()
}

// ListAllowedForProject returns global + project-specific images
func (s *ImageService) ListAllowedForProject(projectID uint) ([]image.AllowedImage, error) {
	return s.repo.FindAllowedImagesForProject(projectID)
}

// AddProjectImage directly adds an image to a project (for project managers)
func (s *ImageService) AddProjectImage(userID, projectID uint, name, tag string) (*image.AllowedImage, error) {
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		return nil, fmt.Errorf("invalid image format: %s", warn)
	}

	img := &image.AllowedImage{
		Name:      name,
		Tag:       tag,
		ProjectID: &projectID,
		IsGlobal:  false,
		CreatedBy: userID,
	}

	if err := s.repo.CreateAllowed(img); err != nil {
		return nil, err
	}
	return img, nil
}

// ValidateImageForProject checks if image is allowed for a project
func (s *ImageService) ValidateImageForProject(name, tag string, projectID uint) (bool, error) {
	return s.repo.ValidateImageForProject(name, tag, projectID)
}

func (s *ImageService) PullImage(name, tag string) error {
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		log.Printf("[image-validate] warning on pull: %s", warn)
	}

	fullImage := fmt.Sprintf("%s:%s", name, tag)
	ttl := int32(300) // Clean up job 5 minutes after completion

	// Create a Job to pull the image
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "image-puller-",
			Namespace:    "default", // Using default namespace for admin tasks
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:            "puller",
							Image:           fullImage,
							ImagePullPolicy: corev1.PullAlways, // Force pull
							// We try to run a harmless command.
							// If the image lacks sh, it will fail, but the image will be pulled.
							Command: []string{"/bin/sh", "-c", "echo Image pulled successfully"},
						},
					},
				},
			},
		},
	}

	_, err := k8s.Clientset.BatchV1().Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create image pull job: %v", err)
		return err
	}

	log.Printf("Created job to pull image: %s", fullImage)
	return nil
}

// validateNameAndTag performs lightweight format checks; returns warning string but does not block.
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

func (s *ImageService) RemoveProjectImage(projectID, imageID uint) error {
	img, err := s.repo.FindAllowedByID(imageID)
	if err != nil {
		return err
	}
	if img.ProjectID == nil || *img.ProjectID != projectID {
		return fmt.Errorf("image does not belong to this project")
	}
	return s.repo.DeleteAllowedImage(imageID)
}

func (s *ImageService) DeleteAllowedImage(id uint) error {
	return s.repo.DeleteAllowedImage(id)
}
