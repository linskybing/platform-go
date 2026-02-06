package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerProjectRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	projects := r.Group("/projects")
	{
		projects.GET("", h.Project.GetProjects)
		projects.POST("", h.Project.CreateProject)
		projects.GET("/by-user", h.Project.GetProjectsByUser)
		projects.GET("/:id", h.Project.GetProjectByID)
		projects.PUT("/:id", h.Project.UpdateProject)
		projects.DELETE("/:id", h.Project.DeleteProject)
	}
}
