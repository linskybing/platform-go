package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerImageRoutes(r *gin.RouterGroup, h *handlers.Handlers, auth *middleware.AuthMiddleware) {
	images := r.Group("/images")
	{
		// Temporarily disable job-related image endpoints until scheduler is externalized
		// images.GET("/pull-active", auth.Admin(), h.Image.GetActivePullJobs)
		// images.GET("/pull-failed", auth.Admin(), h.Image.GetFailedPullJobs)
		// images.POST("/pull", auth.Admin(), h.Image.PullImage)
	}

	// Global image requests endpoints (admin only)
	imageRequests := r.Group("/image-requests")
	{
		imageRequests.GET("", auth.Admin(), h.Image.ListAllImageRequests)
		imageRequests.PUT("/:id/approve", auth.Admin(), h.Image.ApproveImageRequest)
		imageRequests.PUT("/:id/reject", auth.Admin(), h.Image.RejectImageRequest)
	}

	// Avoid unused variable when endpoints are disabled
	_ = images // images group intentionally unused while job endpoints are disabled
}
