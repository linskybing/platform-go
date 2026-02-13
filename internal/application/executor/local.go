package executor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	"gorm.io/datatypes"
)

// LocalExecutor executes jobs directly on Kubernetes without external scheduler
type LocalExecutor struct {
	repos *repository.Repos
}

// NewLocalExecutor creates a new LocalExecutor
func NewLocalExecutor(repos *repository.Repos) Executor {
	return &LocalExecutor{
		repos: repos,
	}
}

// Submit submits a job for immediate execution on Kubernetes
func (e *LocalExecutor) Submit(ctx context.Context, req *SubmitRequest) (*SubmitResult, error) {
	// Create job record with status "running"
	now := time.Now()
	j := &job.Job{
		ID:           req.JobID,
		ConfigFileID: req.ConfigFileID,
		ProjectID:    req.ProjectID,
		Namespace:    req.Namespace,
		UserID:       req.UserID,
		Status:       string(JobStatusRunning),
		SubmitType:   string(req.SubmitType),
		QueueName:    req.QueueName,
		Priority:     req.Priority,
		SubmittedAt:  now,
		StartedAt:    &now,
	}

	if err := e.repos.Job.Create(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to create job record: %w", err)
	}

	// Deploy resources to Kubernetes
	for _, payload := range req.Resources {
		if err := k8s.CreateByJson(datatypes.JSON(payload.JSONData), req.Namespace); err != nil {
			// Update job status to failed
			errMsg := fmt.Sprintf("failed to create resource %s: %v", payload.Name, err)
			_ = e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusFailed), &errMsg)
			return nil, fmt.Errorf("failed to create resource in k8s: %w", err)
		}
	}

	return &SubmitResult{
		JobID:  req.JobID,
		Status: JobStatusRunning,
	}, nil
}

// Cancel cancels a running job by deleting its K8s resources and updating status.
func (e *LocalExecutor) Cancel(ctx context.Context, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("job ID is required")
	}

	jobRecord, err := e.repos.Job.Get(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	status := JobStatus(jobRecord.Status)
	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCancelled {
		return fmt.Errorf("cannot cancel job in %s state", status)
	}

	if jobRecord.Namespace != "" {
		switch jobRecord.SubmitType {
		case string(SubmitTypeWorkflow):
			if err := k8s.DeleteWorkflow(ctx, jobRecord.Namespace, jobID); err != nil {
				slog.Warn("failed to delete workflow during cancel", "job_id", jobID, "namespace", jobRecord.Namespace, "error", err)
			}
		default:
			if err := k8s.DeleteJob(ctx, jobRecord.Namespace, jobID); err != nil {
				slog.Warn("failed to delete k8s job during cancel", "job_id", jobID, "namespace", jobRecord.Namespace, "error", err)
			}
		}
	}

	return e.repos.Job.UpdateStatus(ctx, jobID, string(JobStatusCancelled), nil)
}

// Status retrieves the job status from database
func (e *LocalExecutor) Status(ctx context.Context, jobID string) (JobStatus, error) {
	j, err := e.repos.Job.Get(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("failed to get job: %w", err)
	}
	currentStatus := JobStatus(j.Status)
	if strings.ToLower(j.SubmitType) != string(SubmitTypeWorkflow) {
		return currentStatus, nil
	}

	phase, err := k8s.GetWorkflowPhase(ctx, j.Namespace, jobID)
	if err != nil {
		return currentStatus, err
	}
	if phase == "" {
		return currentStatus, nil
	}
	status := mapWorkflowPhase(phase)
	_ = e.repos.Job.UpdateStatus(ctx, jobID, string(status), nil)
	return status, nil
}

func mapWorkflowPhase(phase string) JobStatus {
	switch strings.ToLower(phase) {
	case "pending":
		return JobStatusQueued
	case "running":
		return JobStatusRunning
	case "succeeded":
		return JobStatusCompleted
	case "failed", "error":
		return JobStatusFailed
	default:
		return JobStatusQueued
	}
}
