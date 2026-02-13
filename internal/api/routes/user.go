package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerUserRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	users := r.Group("/users")
	{
		// Admin User Management (usually) - access control checked inside handler or middleware
		// For now registering basic CRUD.
		users.GET("/", h.User.GetUsers)
		users.GET("/paging", h.User.ListUsersPaging)
		users.GET("/:id", h.User.GetUserByID)
		users.PUT("/:id", h.User.UpdateUser)
		users.DELETE("/:id", h.User.DeleteUser)
		users.GET("/:id/settings", h.User.GetUserSettings)
		users.PUT("/:id/settings", h.User.UpdateUserSettings)
	}
}
