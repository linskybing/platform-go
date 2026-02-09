package image

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"log/slog"

	"github.com/linskybing/platform-go/internal/constants"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
		TagID:    tagEntity.ID,
		IsPulled: true,
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
