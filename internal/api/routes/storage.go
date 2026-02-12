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

			// List PVC bindings for a project - group member access
			pvcBinding.GET("/project/:project_id", am.GroupMember(middleware.FromProjectIDParamName("project_id")), h.PVCBinding.ListBindings)

			// Delete PVC binding by ID
			pvcBinding.DELETE("/:binding_id", h.PVCBinding.DeleteBindingByID)

			// Delete PVC binding by project and pvc name - group manager access
			// use explicit 'project' segment to avoid conflict with '/:binding_id'
			pvcBinding.DELETE("/project/:project_id/:pvc_name", am.GroupManager(middleware.FromProjectIDParamName("project_id")), h.PVCBinding.DeleteBinding)
		}

		// FileBrowser access routes
		filebrowser := k8s.Group("/filebrowser")
		{
			// Get FileBrowser access - group member access
			// Note: Handler should validate user has access to requested project/storage
			filebrowser.POST("/access", h.FileBrowser.GetAccess)
		}
	}

	// Storage permission routes
	storage := auth.Group("/storage")
	{
		// Group storage routes
		// List storages for a group - group members can view
		storage.GET("/group/:id", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.ListGroupStorages)

		// List current user's accessible storages
		storage.GET("/my-storages", h.GroupStorage.GetMyGroupStorages)

		// Create group storage - group admin
		storage.POST("/:id/storage", am.GroupAdmin(middleware.FromGroupIDParam()), h.GroupStorage.CreateGroupStorage)

		// Delete group storage - group admin
		storage.DELETE("/:id/storage/:pvcId", am.GroupAdmin(middleware.FromGroupIDParam()), h.GroupStorage.DeleteGroupStorage)

		// Start/stop FileBrowser for a group PVC
		storage.POST("/:id/storage/:pvcId/start", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.StartFileBrowser)
		storage.DELETE("/:id/storage/:pvcId/stop", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.StopFileBrowser)

		permissions := storage.Group("/permissions")
		{
			// Set permission - group admin only (via group_id in payload)
			permissions.POST("", am.GroupAdmin(middleware.FromGroupIDInPayload()), h.StoragePerm.SetPermission)
			permissions.POST("/batch", am.GroupAdmin(middleware.FromGroupIDInPayload()), h.StoragePerm.BatchSetPermissions)
			permissions.GET("/group/:group_id/pvc/:pvc_id", am.GroupMember(middleware.FromGroupIDParamName("group_id")), h.StoragePerm.GetUserPermission)
			permissions.GET("/group/:group_id/pvc/:pvc_id/list", am.GroupMember(middleware.FromGroupIDParamName("group_id")), h.StoragePerm.ListPVCPermissions)
		}

		// Access policy routes
		storage.POST("/policies", am.GroupAdmin(middleware.FromGroupIDInPayload()), h.StoragePerm.SetAccessPolicy)
	}
}
