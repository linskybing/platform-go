package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerStorageRoutes registers Kubernetes storage and PVC binding routes.
// All routes are protected by JWT middleware with role-based authorization.
//
// Authorization levels:
// - PVC binding: Group manager (can bind PVC to projects they manage)
// - FileBrowser access: Group member (can access file browser for their projects)
func registerStorageRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	k8s := auth.Group("/k8s")
	{
		// PVC binding routes
		pvcBinding := k8s.Group("/pvc-binding")
		{
			// Create PVC binding - group manager access (via project_id in payload)
			pvcBinding.POST("", am.GroupManager(middleware.FromProjectIDInPayload()), h.PVCBinding.CreateBinding)
		}

		// FileBrowser access routes
		filebrowser := k8s.Group("/filebrowser")
		{
			// Get FileBrowser access - group member access
			// Note: Handler should validate user has access to requested project/storage
			filebrowser.GET("/access", h.FileBrowser.GetAccess)
		}
	}
}
