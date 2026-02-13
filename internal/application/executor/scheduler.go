package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	k8stypes "github.com/linskybing/platform-go/pkg/k8s/types"
	"gorm.io/datatypes"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchedulerExecutor submits jobs to an external scheduler (flash-sched)
// This is a stub implementation for future integration
type SchedulerExecutor struct {
	repos          *repository.Repos
	flashJobClient *k8s.FlashJobClient
	schedulerName  string
}

// NewSchedulerExecutor creates a new SchedulerExecutor
func NewSchedulerExecutor(repos *repository.Repos, flashJobClient *k8s.FlashJobClient, schedulerName string) Executor {
	return &SchedulerExecutor{
		repos:          repos,
		flashJobClient: flashJobClient,
		schedulerName:  schedulerName,
	}
}

// Submit submits a job to the external scheduler
// Expected contract:
// 1. POST to flash-sched API with queue_name, priority, resource definitions
// 2. Scheduler queues the job and returns job ID
// 3. Job record is created in database with status "queued"
// 4. Scheduler executes job when resources are available
// 5. Status updates are polled or pushed via webhook
func (e *SchedulerExecutor) Submit(ctx context.Context, req *SubmitRequest) (*SubmitResult, error) {
	if req == nil {
		return nil, fmt.Errorf("submit request is nil")
	}
	if e.flashJobClient == nil {
		return nil, fmt.Errorf("flashjob client not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if req.JobID == "" {
		return nil, fmt.Errorf("job ID is required")
	}
	if req.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	queueName := req.QueueName
	if queueName == "" {
		queueName = "default-batch"
	}

	jobRecord := &job.Job{
		ID:           req.JobID,
		ConfigFileID: req.ConfigFileID,
		ProjectID:    req.ProjectID,
		Namespace:    req.Namespace,
		UserID:       req.UserID,
		Status:       string(JobStatusQueued),
		SubmitType:   string(req.SubmitType),
		QueueName:    queueName,
		Priority:     req.Priority,
		SubmittedAt:  time.Now(),
	}
	if err := e.repos.Job.Create(ctx, jobRecord); err != nil {
		return nil, fmt.Errorf("failed to create job record: %w", err)
	}

	templates, extraResources, err := buildPodTemplates(req.Resources)
	if err != nil {
		errMsg := err.Error()
		_ = e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusFailed), &errMsg)
		return nil, err
	}
	if len(templates) == 0 {
		errMsg := "no pod templates found for flashjob"
		_ = e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusFailed), &errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	applySchedulingMetadata(templates, queueName, e.schedulerName)

	for _, payload := range extraResources {
		if err := k8s.CreateByJson(datatypes.JSON(payload.JSONData), req.Namespace); err != nil {
			errMsg := fmt.Sprintf("failed to create resource %s: %v", payload.Name, err)
			_ = e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusFailed), &errMsg)
			return nil, fmt.Errorf("failed to create resource in k8s: %w", err)
		}
	}

	flashJob := &k8stypes.FlashJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.flash-sched.io/v1alpha1",
			Kind:       "FlashJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.JobID,
			Namespace: req.Namespace,
			Labels: map[string]string{
				"platform.job-id":        req.JobID,
				"platform.project-id":    req.ProjectID,
				"platform.configfile-id": req.ConfigFileID,
				"platform.user-id":       req.UserID,
				"platform.username":      req.Username,
			},
		},
		Spec: k8stypes.FlashJobSpec{
			MinAvailable: 1,
			Tasks:        templates,
		},
	}

	if _, err := e.flashJobClient.Create(ctx, req.Namespace, flashJob); err != nil {
		errMsg := err.Error()
		_ = e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusFailed), &errMsg)
		return nil, fmt.Errorf("create flashjob: %w", err)
	}

	return &SubmitResult{JobID: req.JobID, Status: JobStatusQueued}, nil
}

// Cancel requests the scheduler to cancel or evict a job
// Expected contract:
// 1. POST to flash-sched cancel endpoint with job ID
// 2. Scheduler stops/evicts the job if running
// 3. Job status is updated to "cancelled"
func (e *SchedulerExecutor) Cancel(ctx context.Context, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("job ID is required")
	}
	if e.flashJobClient == nil {
		return fmt.Errorf("flashjob client not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	jobRecord, err := e.repos.Job.Get(ctx, jobID)
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}
	if err := e.flashJobClient.Delete(ctx, jobRecord.Namespace, jobID); err != nil {
		return fmt.Errorf("delete flashjob: %w", err)
	}
	return e.repos.Job.UpdateStatus(ctx, jobID, string(JobStatusCancelled), nil)
}

// Status queries the scheduler for job lifecycle status
// Expected contract:
// 1. GET from flash-sched status endpoint with job ID
// 2. Returns current status (queued, running, completed, failed, cancelled)
// 3. Job database record is updated with latest status
func (e *SchedulerExecutor) Status(ctx context.Context, jobID string) (JobStatus, error) {
	if jobID == "" {
		return "", fmt.Errorf("job ID is required")
	}
	if e.flashJobClient == nil {
		return "", fmt.Errorf("flashjob client not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	jobRecord, err := e.repos.Job.Get(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("get job: %w", err)
	}
	flashJob, err := e.flashJobClient.Get(ctx, jobRecord.Namespace, jobID)
	if err != nil {
		return "", fmt.Errorf("get flashjob: %w", err)
	}
	status := mapFlashJobPhase(flashJob.Status.Phase)
	_ = e.repos.Job.UpdateStatus(ctx, jobID, string(status), nil)
	return status, nil
}

func mapFlashJobPhase(phase k8stypes.FlashJobPhase) JobStatus {
	switch phase {
	case k8stypes.FlashJobRunning:
		return JobStatusRunning
	case k8stypes.FlashJobSucceeded:
		return JobStatusCompleted
	case k8stypes.FlashJobFailed:
		return JobStatusFailed
	case k8stypes.FlashJobPending:
		return JobStatusQueued
	default:
		return JobStatusQueued
	}
}

func buildPodTemplates(resources []ResourcePayload) ([]corev1.PodTemplateSpec, []ResourcePayload, error) {
	templates := make([]corev1.PodTemplateSpec, 0)
	extra := make([]ResourcePayload, 0)
	for _, payload := range resources {
		var meta struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(payload.JSONData, &meta); err != nil {
			return nil, nil, fmt.Errorf("parse resource %s: %w", payload.Name, err)
		}
		switch meta.Kind {
		case "Pod":
			var pod corev1.Pod
			if err := json.Unmarshal(payload.JSONData, &pod); err != nil {
				return nil, nil, fmt.Errorf("parse pod %s: %w", payload.Name, err)
			}
			templates = append(templates, corev1.PodTemplateSpec{
				ObjectMeta: pod.ObjectMeta,
				Spec:       pod.Spec,
			})
		case "Job":
			var job batchv1.Job
			if err := json.Unmarshal(payload.JSONData, &job); err != nil {
				return nil, nil, fmt.Errorf("parse job %s: %w", payload.Name, err)
			}
			templates = append(templates, job.Spec.Template)
		default:
			extra = append(extra, payload)
		}
	}
	return templates, extra, nil
}

func applySchedulingMetadata(templates []corev1.PodTemplateSpec, queueName, schedulerName string) {
	for i := range templates {
		if templates[i].Annotations == nil {
			templates[i].Annotations = make(map[string]string)
		}
		if queueName != "" {
			templates[i].Annotations["scheduling.flash-sched.io/queue-name"] = queueName
		}
		if schedulerName != "" {
			templates[i].Spec.SchedulerName = schedulerName
		}
	}
}
