package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerUserRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	users := r.Group("/users")
	{
		// Self management
		users.POST("/logout", h.User.Logout)

		// Admin User Management (usually) - access control checked inside handler or middleware
		// For now registering basic CRUD.
		users.GET("/:id", h.User.GetUserByID)
		users.PUT("/:id", h.User.UpdateUser)
		users.DELETE("/:id", h.User.DeleteUser)
		// users.GET("/", h.User.ListUsers) // If implemented
	}

	userGroups := r.Group("/user-groups")
	{
		userGroups.POST("", h.UserGroup.AddUserToGroup)
		userGroups.DELETE("", h.UserGroup.RemoveUserFromGroup)
		userGroups.POST("/role", h.UserGroup.UpdateUserRole)
		userGroups.GET("/:group_id/members", h.UserGroup.GetGroupMembers) // Adjusted path if needed
	}
}
