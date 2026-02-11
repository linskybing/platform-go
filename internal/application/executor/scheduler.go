package executor

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/repository"
)

// SchedulerExecutor submits jobs to an external scheduler (flash-sched)
// This is a stub implementation for future integration
type SchedulerExecutor struct {
	repos         *repository.Repos
	schedulerURL  string
	schedulerAuth string
}

// NewSchedulerExecutor creates a new SchedulerExecutor
func NewSchedulerExecutor(repos *repository.Repos, schedulerURL, schedulerAuth string) Executor {
	return &SchedulerExecutor{
		repos:         repos,
		schedulerURL:  schedulerURL,
		schedulerAuth: schedulerAuth,
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
	return nil, fmt.Errorf("SchedulerExecutor not implemented: flash-sched integration pending")
}

// Cancel requests the scheduler to cancel or evict a job
// Expected contract:
// 1. POST to flash-sched cancel endpoint with job ID
// 2. Scheduler stops/evicts the job if running
// 3. Job status is updated to "cancelled"
func (e *SchedulerExecutor) Cancel(ctx context.Context, jobID string) error {
	return fmt.Errorf("SchedulerExecutor not implemented: flash-sched integration pending")
}

// Status queries the scheduler for job lifecycle status
// Expected contract:
// 1. GET from flash-sched status endpoint with job ID
// 2. Returns current status (queued, running, completed, failed, cancelled)
// 3. Job database record is updated with latest status
func (e *SchedulerExecutor) Status(ctx context.Context, jobID string) (JobStatus, error) {
	return "", fmt.Errorf("SchedulerExecutor not implemented: flash-sched integration pending")
}
