package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/repository"
)

func registerProjectRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware, repos *repository.Repos) {
	projects := r.Group("/projects")
	{
		projects.GET("", h.Project.GetProjects)
		projects.GET("/by-user", h.Project.GetProjectsByUser)
		projects.GET("/:id", h.Project.GetProjectByID)
		projects.GET("/:id/config-files", h.ConfigFile.ListConfigFilesByProjectIDHandler)
		projects.POST("", authMw.Admin(), h.Project.CreateProject)
		projects.PUT("/:id", authMw.GroupManager(middleware.FromProjectIDParam(repos)), h.Project.UpdateProject)
		projects.DELETE("/:id", authMw.GroupManager(middleware.FromProjectIDParam(repos)), h.Project.DeleteProject)

		// Project-level image management (for project managers)
		projects.GET("/:id/images", authMw.GroupMember(middleware.FromProjectIDInPayload()), h.Image.ListAllowed)
		projects.POST("/:id/images", authMw.GroupManager(middleware.FromProjectIDInPayload()), h.Image.AddProjectImage)
		projects.DELETE("/:id/images/:image_id", authMw.GroupManager(middleware.FromProjectIDInPayload()), h.Image.RemoveProjectImage)
		// Project-scoped image requests (list requests for a specific project)
		projects.GET("/:id/image-requests", authMw.GroupMember(middleware.FromProjectIDInPayload()), h.Image.ListRequestsByProject)
	}
}
