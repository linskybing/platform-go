package executor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	k8stypes "github.com/linskybing/platform-go/pkg/k8s/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// reconcilerInitialBackoff is the initial backoff duration on watch failures.
	reconcilerInitialBackoff = 2 * time.Second
	// reconcilerMaxBackoff is the maximum backoff duration on repeated failures.
	reconcilerMaxBackoff = 60 * time.Second
)

// StartFlashJobReconciler watches FlashJob CRDs and updates job status in the DB.
// On watch reconnect, it performs a full resync by listing all FlashJobs.
func StartFlashJobReconciler(ctx context.Context, repos *repository.Repos, client *k8s.FlashJobClient) {
	if repos == nil || client == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	go func() {
		backoff := reconcilerInitialBackoff
		for {
			if ctx.Err() != nil {
				return
			}

			// Resync: list all FlashJobs and update DB state
			resyncFlashJobs(ctx, repos, client)

			watcher, err := client.Watch(ctx, "", metav1.ListOptions{})
			if err != nil {
				slog.Warn("flashjob watch failed", "error", err)
				time.Sleep(backoff)
				backoff = min(backoff*2, reconcilerMaxBackoff)
				continue
			}
			backoff = reconcilerInitialBackoff

			err = consumeFlashJobEvents(ctx, watcher, repos)
			if err != nil {
				slog.Warn("flashjob watch ended", "error", err)
				time.Sleep(backoff)
			}
		}
	}()
}

// resyncFlashJobs lists all FlashJobs and updates their status in the DB.
func resyncFlashJobs(ctx context.Context, repos *repository.Repos, client *k8s.FlashJobClient) {
	jobList, err := client.List(ctx, "", metav1.ListOptions{})
	if err != nil {
		slog.Warn("flashjob resync list failed", "error", err)
		return
	}
	if jobList == nil {
		return
	}

	for _, fj := range jobList.Items {
		status := mapFlashJobPhase(fj.Status.Phase)
		if err := repos.Job.UpdateStatus(ctx, fj.Name, string(status), nil); err != nil {
			slog.Debug("flashjob resync update skipped", "job_id", fj.Name, "error", err)
		}
	}
	slog.Info("flashjob resync completed", "count", len(jobList.Items))
}

func consumeFlashJobEvents(ctx context.Context, watcher watch.Interface, repos *repository.Repos) error {
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			job, err := flashJobFromEvent(event)
			if err != nil {
				slog.Warn("flashjob event conversion failed", "error", err)
				continue
			}
			if job == nil {
				continue
			}
			status := mapFlashJobPhase(job.Status.Phase)
			if event.Type == watch.Deleted {
				status = JobStatusCancelled
			}
			if err := repos.Job.UpdateStatus(ctx, job.Name, string(status), nil); err != nil {
				slog.Warn("flashjob status update failed", "job_id", job.Name, "error", err)
			}
		}
	}
}

func flashJobFromEvent(event watch.Event) (*k8stypes.FlashJob, error) {
	if event.Object == nil {
		return nil, nil
	}

	obj, ok := event.Object.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("unsupported event object type")
	}

	var job k8stypes.FlashJob
	if err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
