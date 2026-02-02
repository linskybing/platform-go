package cron

import (
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/application"
)

func StartCleanupTask(auditService *application.AuditService) {
	go func() {
		slog.Info("starting background cleanup task", "retention_days", 30)

		// Run immediately on startup
		if err := auditService.CleanupOldLogs(30); err != nil {
			slog.Error("failed to cleanup old audit logs", "error", err)
		}

		// Then run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			slog.Info("running scheduled audit log cleanup")
			if err := auditService.CleanupOldLogs(30); err != nil {
				slog.Error("failed to cleanup old audit logs", "error", err)
			} else {
				slog.Info("audit log cleanup completed successfully")
			}
		}
	}()
}
