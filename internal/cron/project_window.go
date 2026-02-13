package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/repository"
)

func StartProjectScheduleEnforcer(repos *repository.Repos, exec executor.Executor, interval time.Duration) {
	if repos == nil || repos.Project == nil || repos.Job == nil {
		slog.Warn("project schedule enforcer skipped: repos not configured")
		return
	}
	if exec == nil {
		slog.Warn("project schedule enforcer skipped: executor not configured")
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := enforceProjectSchedules(repos, exec); err != nil {
				slog.Warn("project schedule enforcement failed", "error", err)
			}
			<-ticker.C
		}
	}()
}

func enforceProjectSchedules(repos *repository.Repos, exec executor.Executor) error {
	ctx := context.Background()
	projects, err := repos.Project.ListProjects()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return nil
	}

	now := time.Now()
	for _, proj := range projects {
		allowed, err := proj.IsTimeAllowed(now)
		if err != nil {
			slog.Warn("invalid project schedule", "project_id", proj.PID, "error", err)
			continue
		}
		if allowed {
			continue
		}
		jobs, err := repos.Job.ListByProjectAndStatuses(ctx, proj.PID, []string{
			string(executor.JobStatusQueued),
			string(executor.JobStatusSubmitted),
			string(executor.JobStatusRunning),
		})
		if err != nil {
			slog.Warn("failed to list jobs for project schedule enforcement", "project_id", proj.PID, "error", err)
			continue
		}
		for _, job := range jobs {
			if err := exec.Cancel(ctx, job.ID); err != nil {
				slog.Warn("failed to cancel job outside schedule", "job_id", job.ID, "project_id", proj.PID, "error", err)
			}
		}
	}
	return nil
}
