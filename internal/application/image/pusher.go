package image

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/constants"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/pkg/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ImageService) PullImageAsync(name, tag string) (string, error) {
	if warn := s.validateNameAndTag(name, tag); warn != "" {
		slog.Warn("image validation warning on pull",
			"image", name,
			"tag", tag,
			"warning", warn)
	}

	// --- [修正] 映像檔名稱正規化邏輯 ---
	// 解決 codercom/code-server 變成 code-server:latest 或找不到 Registry 的問題
	normalizedName := name
	parts := strings.Split(name, "/")

	// 檢查第一部分是否包含 Domain 特徵 (. 或 :) 或為 localhost
	hasDomain := strings.Contains(parts[0], ".") ||
		strings.Contains(parts[0], ":") ||
		parts[0] == "localhost"

	if !hasDomain {
		if len(parts) == 1 {
			// Case: "nginx" -> "docker.io/library/nginx"
			normalizedName = "docker.io/library/" + name
		} else {
			// Case: "codercom/code-server" -> "docker.io/codercom/code-server"
			normalizedName = "docker.io/" + name
		}
	}

	// 這是用於 K8s InitContainer 拉取以及 crane copy 來源的完整位址
	fullImage := fmt.Sprintf("%s:%s", normalizedName, tag)

	// 這是推送到內部 Harbor 的位址 (維持原始路徑結構)
	harborImage := fmt.Sprintf("%s%s:%s", config.HarborPrivatePrefix, name, tag)
	// --------------------------------

	ttl := int32(300)

	k8sJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "image-puller-",
			Namespace:    "default",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "default",
					InitContainers: []corev1.Container{
						{
							Name:            "pull-source",
							Image:           fullImage,
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/bin/sh", "-c", "echo 'Image pulled successfully'"},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "push-to-harbor",
							Image:           "gcr.io/go-containerregistry/crane:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"crane", "copy", fullImage, harborImage, "--insecure"},
							Env: []corev1.EnvVar{
								{
									Name:  "DOCKER_CONFIG",
									Value: "/kaniko/.docker",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "docker-config",
									MountPath: "/kaniko/.docker",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "docker-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "harbor-regcred",
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: "config.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

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

func (s *ImageService) markImageAsPulled(name, tag string) {
	parts := strings.Split(name, "/")
	var namespace, repoName string
	if len(parts) >= 2 {
		namespace = parts[0]
		repoName = strings.Join(parts[1:], "/")
	} else {
		namespace = "library"
		repoName = name
	}

	repo := &image.ContainerRepository{
		Namespace: namespace,
		Name:      repoName,
		FullName:  name,
	}
	if err := s.repo.FindOrCreateRepository(repo); err != nil {
		slog.Error("failed to find or create repository for status update",
			"name", name,
			"namespace", namespace,
			"error", err)
		return
	}

	tagEntity := &image.ContainerTag{
		RepositoryID: repo.ID,
		Name:         tag,
	}
	if err := s.repo.FindOrCreateTag(tagEntity); err != nil {
		slog.Error("failed to find or create tag for status update",
			"repo_id", repo.ID,
			"tag", tag,
			"error", err)
		return
	}

	status := &image.ClusterImageStatus{
		TagID:        tagEntity.ID,
		IsPulled:     true,
		LastPulledAt: ptrTime(time.Now()),
	}

	if err := s.repo.UpdateClusterStatus(status); err != nil {
		slog.Error("failed to update cluster image status",
			"tag_id", tagEntity.ID,
			"image", name,
			"tag", tag,
			"error", err)
	} else {
		slog.Info("cluster image status updated",
			"image", name,
			"tag", tag)
	}
}

func (s *ImageService) getPodLogsForJob(jobName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), constants.K8sQuickOpTimeout*time.Second)
	defer cancel()

	labelSelector := fmt.Sprintf("job-name=%s", jobName)

	pods, err := k8s.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		slog.Debug("error listing pods for logs",
			"job_name", jobName,
			"error", err)
		return fmt.Sprintf("Error listing pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return "No pods found for this job"
	}

	var logBuilder strings.Builder
	for i := range pods.Items {
		pod := &pods.Items[i]
		logBuilder.WriteString(fmt.Sprintf("=== Pod: %s ===\n", pod.Name))

		req := k8s.Clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			TailLines: func() *int64 { i := int64(100); return &i }(),
		})

		stream, err := req.Stream(ctx)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("Error getting logs: %v\n", err))
			continue
		}

		data, err := io.ReadAll(stream)
		_ = stream.Close()
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("Error reading logs: %v\n", err))
			continue
		}

		logBuilder.Write(data)
		logBuilder.WriteString("\n")
	}

	return logBuilder.String()
}
