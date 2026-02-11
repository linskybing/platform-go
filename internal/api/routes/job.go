package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerJobRoutes registers all job-related routes
func registerJobRoutes(r *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	jobs := r.Group("/api/jobs")
	{
		jobs.GET("", h.Job.ListJobs)
		jobs.GET("/:id", h.Job.GetJob)
		jobs.POST("/:id/cancel", h.Job.CancelJob)
	}
}
