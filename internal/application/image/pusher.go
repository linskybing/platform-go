package image

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/constants"
	"github.com/linskybing/platform-go/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ImageService) PullImageAsync(name, tag string) (string, error) {
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		slog.Warn("image validation warning on pull",
			"image", name,
			"tag", tag,
			"warning", warn)
	}
	// Build k8s job and canonical image names using extracted builder
	k8sJob, fullImage, harborImage := BuildPullJob(name, tag)

	// Create job with context from background
	ctx, cancel := context.WithTimeout(context.Background(), constants.K8sStandardOpTimeout*time.Second)
	defer cancel()

	createdJob, err := k8s.Clientset.BatchV1().Jobs("default").Create(ctx, k8sJob, metav1.CreateOptions{})
	if err != nil {
		slog.Error("failed to create image pull job",
			"image", fullImage,
			"error", err)
		return "", fmt.Errorf("failed to create image pull job: %w", err)
	}

	jobID := createdJob.Name
	pullTracker.AddJob(jobID, name, tag)
	pullTracker.UpdateJob(jobID, "pulling", 10, "Starting image pull...")

	go s.monitorPullJob(jobID, name, tag)

	slog.Info("image pull job created",
		"job_id", jobID,
		"image", fullImage,
		"harbor_image", harborImage)
	return jobID, nil
}
