package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerUserGroupsRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	userGroups := r.Group("/user-groups")
	{
		userGroups.POST("", authMw.GroupAdmin(middleware.FromGroupIDInPayload()), h.UserGroup.AddUserToGroup)
		userGroups.PUT("", authMw.GroupAdmin(middleware.FromGroupIDInPayload()), h.UserGroup.UpdateUserGroup)
		userGroups.DELETE("", authMw.GroupAdmin(middleware.FromGroupIDInPayload()), h.UserGroup.RemoveUserFromGroup)

		userGroups.GET("", authMw.Admin(), h.UserGroup.GetUserGroup)
		userGroups.GET("/:group_id/members", h.UserGroup.GetGroupMembers) // Adjusted path if needed
		userGroups.GET("/by-group", h.UserGroup.GetUserGroupsByGID)
		userGroups.GET("/by-user", h.UserGroup.GetUserGroupsByUID)
	}
}
