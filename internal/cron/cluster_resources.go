package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/application/cluster"
)

func StartClusterResourceCollector(service *cluster.ClusterService) {
	if service == nil {
		return
	}
	go func() {
		slog.Info("starting cluster resource collector")

		ctx := context.Background()
		if err := service.RefreshCache(ctx); err != nil {
			slog.Warn("cluster resource collection failed", "error", err)
		}

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := service.RefreshCache(ctx); err != nil {
				slog.Warn("cluster resource collection failed", "error", err)
			}
		}
	}()
}
