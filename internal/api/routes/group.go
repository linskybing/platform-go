package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerGroupRoutes registers group management routes.
// All routes are protected by JWT middleware with role-based authorization.
//
// Authorization levels:
// - GET (list/view): Group member access (any member can view their groups)
// - POST (create): Admin only (only super admin can create groups)
// - PUT (update): Group admin (admin of the specific group)
// - DELETE: Group admin (admin of the specific group)
func registerGroupRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	groups := auth.Group("/groups")
	{
		// View groups - any authenticated user can see their groups
		groups.GET("", h.Group.GetGroups)

		// View specific group - members can view their group
		groups.GET("/:id", am.GroupMember(middleware.FromGroupIDParam()), h.Group.GetGroupByID)

		// Create group - admin only
		groups.POST("", am.Admin(), h.Group.CreateGroup)

		// Update group - group admin only
		groups.PUT("/:id", am.GroupAdmin(middleware.FromGroupIDParam()), h.Group.UpdateGroup)

		// Delete group - group admin only
		groups.DELETE("/:id", am.GroupAdmin(middleware.FromGroupIDParam()), h.Group.DeleteGroup)

		// Group storage routes
		// List storages for a group - group members can view
		groups.GET("/:id/storage", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.ListGroupStorages)

		// List current user's accessible storages
		groups.GET("/my-storages", h.GroupStorage.GetMyGroupStorages)

		// Create group storage - group admin
		groups.POST("/:id/storage", am.GroupAdmin(middleware.FromGroupIDParam()), h.GroupStorage.CreateGroupStorage)

		// Delete group storage - group admin
		groups.DELETE("/:id/storage/:pvcId", am.GroupAdmin(middleware.FromGroupIDParam()), h.GroupStorage.DeleteGroupStorage)

		// Start/stop FileBrowser for a group PVC
		groups.POST("/:id/storage/:pvcId/start", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.StartFileBrowser)
		groups.DELETE("/:id/storage/:pvcId/stop", am.GroupMember(middleware.FromGroupIDParam()), h.GroupStorage.StopFileBrowser)
	}
}
