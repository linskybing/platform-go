package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerAdminStorageRoutes registers admin-only storage management routes
// Protected by Admin() middleware - only users with admin role can access
func registerAdminStorageRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	admin := auth.Group("/admin/user-storage")
	admin.Use(am.Admin())
	{
		// User storage management endpoints
		admin.GET("/:username/status", h.UserStorage.CheckStatus)
		admin.POST("/:username/init", h.UserStorage.Initialize)
		admin.PUT("/:username/expand", h.UserStorage.Expand)
		admin.DELETE("/:username", h.UserStorage.Delete)
	}
}
