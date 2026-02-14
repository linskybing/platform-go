package priority

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/priority/preemptor"
	"github.com/linskybing/platform-go/internal/repository"
)

// PreemptionService handles the execution of preemption decisions.
type PreemptionService struct {
	manager *Manager
	repos   *repository.Repos
}

// NewPreemptionService creates a new service for managing preemption.
func NewPreemptionService(mgr *Manager, repos *repository.Repos) *PreemptionService {
	return &PreemptionService{manager: mgr, repos: repos}
}

// PerformPreemption finds and evicts victim jobs to free up resources.
func (s *PreemptionService) PerformPreemption(ctx context.Context, req preemptor.ResourceRequirement) error {
	decision, err := s.manager.Preempt(ctx, req)
	if err != nil {
		return fmt.Errorf("preemption decision failed: %w", err)
	}

	for _, jobID := range decision.JobsToPreempt {
		if err := s.evictJob(ctx, jobID, decision.Reason); err != nil {
			return fmt.Errorf("failed to evict job %s: %w", jobID, err)
		}
	}
	return nil
}

// evictJob performs the actual eviction of a job from the cluster.
func (s *PreemptionService) evictJob(ctx context.Context, jobID string, reason string) error {
	err := s.repos.Job.UpdateStatus(ctx, jobID, "PREEMPTED", &reason)
	if err != nil {
		return err
	}
	// TODO: Call K8s API to delete the pod/job.
	return nil
}
