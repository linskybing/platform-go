package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/application/gpuusage"
)

func StartGPUUsageCollector(service *gpuusage.GPUUsageService) {
	if service == nil {
		return
	}
	go func() {
		slog.Info("starting gpu usage snapshot collector")

		ctx := context.Background()
		if err := service.CollectSnapshots(ctx); err != nil {
			slog.Warn("gpu usage snapshot failed", "error", err)
		}
		if err := service.ComputeMissingSummaries(ctx); err != nil {
			slog.Warn("gpu usage summary failed", "error", err)
		}

		snapshotTicker := time.NewTicker(30 * time.Second)
		defer snapshotTicker.Stop()
		cleanupTicker := time.NewTicker(24 * time.Hour)
		defer cleanupTicker.Stop()

		for {
			select {
			case <-snapshotTicker.C:
				if err := service.CollectSnapshots(ctx); err != nil {
					slog.Warn("gpu usage snapshot failed", "error", err)
				}
				if err := service.ComputeMissingSummaries(ctx); err != nil {
					slog.Warn("gpu usage summary failed", "error", err)
				}
			case <-cleanupTicker.C:
				if err := service.CleanupSnapshots(ctx, 30*24*time.Hour); err != nil {
					slog.Warn("gpu usage cleanup failed", "error", err)
				}
			}
		}
	}()
}
