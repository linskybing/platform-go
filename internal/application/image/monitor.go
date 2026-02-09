package image

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/linskybing/platform-go/internal/constants"
	"github.com/linskybing/platform-go/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// monitorPullJob observes the K8s Job and updates the pullTracker accordingly.
func (s *ImageService) monitorPullJob(jobID, imageName, imageTag string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.K8sImagePullTimeout*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	maxRetries := 600
	retries := 0

	for range ticker.C {
		retries++
		if retries > maxRetries {
			logs := s.getPodLogsForJob(jobID)
			errMsg := fmt.Sprintf("Job timeout. Logs: %s", logs)
			pullTracker.UpdateJob(jobID, "failed", 0, errMsg)
			pullTracker.RemoveJob(jobID)
			slog.Error("image pull job timeout",
				"job_id", jobID,
				"image", imageName,
				"tag", imageTag,
				"logs", logs)
			return
		}

		k8sJob, err := k8s.Clientset.BatchV1().Jobs("default").Get(ctx, jobID, metav1.GetOptions{})
		if err != nil {
			slog.Debug("error getting image pull job",
				"job_id", jobID,
				"image", imageName,
				"error", err)
			continue
		}

		labelSelector := fmt.Sprintf("job-name=%s", jobID)
		pods, err := k8s.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})

		var statusMsg string
		var progress int

		if err == nil && len(pods.Items) > 0 {
			pod := &pods.Items[0]

			if len(pod.Status.InitContainerStatuses) > 0 {
				initStatus := pod.Status.InitContainerStatuses[0]
				if initStatus.State.Running != nil {
					statusMsg = "Pulling source image..."
					progress = 30
				} else if initStatus.State.Terminated != nil {
					if initStatus.State.Terminated.ExitCode == 0 {
						statusMsg = "Source image pulled, pushing to Harbor..."
						progress = 60
					} else {
						logs := s.getPodLogsForJob(jobID)
						errMsg := fmt.Sprintf("Failed to pull source image. Logs: %s", logs)
						pullTracker.UpdateJob(jobID, "failed", 0, errMsg)
						pullTracker.RemoveJob(jobID)
						return
					}
				} else if initStatus.State.Waiting != nil {
					statusMsg = fmt.Sprintf("Waiting: %s", initStatus.State.Waiting.Reason)
					progress = 10
				}
			}

			if len(pod.Status.ContainerStatuses) > 0 {
				containerStatus := pod.Status.ContainerStatuses[0]
				if containerStatus.State.Running != nil {
					statusMsg = "Pushing to Harbor..."
					progress = 80
				} else if containerStatus.State.Waiting != nil {
					if statusMsg == "" {
						statusMsg = fmt.Sprintf("Waiting: %s", containerStatus.State.Waiting.Reason)
						progress = 50
					}
				}
			}

			switch pod.Status.Phase {
			case corev1.PodPending:
				if statusMsg == "" {
					statusMsg = "Pod pending..."
					progress = 5
				}
			case corev1.PodRunning:
				if statusMsg == "" {
					statusMsg = "Processing..."
					progress = 50
				}
			case corev1.PodFailed:
				logs := s.getPodLogsForJob(jobID)
				errMsg := fmt.Sprintf("Pod failed. Logs: %s", logs)
				pullTracker.UpdateJob(jobID, "failed", 0, errMsg)
				pullTracker.RemoveJob(jobID)
				return
			}
		} else {
			statusMsg = "Initializing..."
			progress = 5
		}

		if k8sJob.Status.Succeeded > 0 {
			pullTracker.UpdateJob(jobID, "completed", 100, "Image pushed to Harbor successfully")

			s.markImageAsPulled(imageName, imageTag)

			pullTracker.RemoveJob(jobID)
			return
		}

		if k8sJob.Status.Failed > 0 {
			logs := s.getPodLogsForJob(jobID)
			errMsg := fmt.Sprintf("Job failed. Logs: %s", logs)
			pullTracker.UpdateJob(jobID, "failed", 0, errMsg)
			pullTracker.RemoveJob(jobID)
			return
		}

		pullTracker.UpdateJob(jobID, "pulling", progress, statusMsg)
	}
}
