package executor

import (
	"context"
	"fmt"
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

	// Update job status to completed
	completedAt := time.Now()
	j.Status = string(JobStatusCompleted)
	j.CompletedAt = &completedAt

	if err := e.repos.Job.UpdateStatus(ctx, req.JobID, string(JobStatusCompleted), nil); err != nil {
		// Log error but don't fail the submission since resources are already created
		return &SubmitResult{
			JobID:  req.JobID,
			Status: JobStatusCompleted,
		}, nil
	}

	return &SubmitResult{
		JobID:  req.JobID,
		Status: JobStatusCompleted,
	}, nil
}

// Cancel is not implemented for local executor (jobs execute immediately)
func (e *LocalExecutor) Cancel(ctx context.Context, jobID string) error {
	return fmt.Errorf("cancel not supported for local executor")
}

// Status retrieves the job status from database
func (e *LocalExecutor) Status(ctx context.Context, jobID string) (JobStatus, error) {
	j, err := e.repos.Job.Get(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("failed to get job: %w", err)
	}
	return JobStatus(j.Status), nil
}
