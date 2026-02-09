package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/repository"
)

// registerConfigFileRoutes registers configuration file management routes.
// All routes are protected by JWT middleware with role-based authorization.
//
// Authorization levels:
// - GET (list/view): Group member (can view config files in their groups)
// - POST (create): Group manager (can create in projects they manage)
// - PUT (update): Group manager (can update files they own)
// - DELETE: Group manager (can delete files they own)
// - Instance operations: Group member (can create/destroy instances)
func registerConfigFileRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware, repos *repository.Repos) {
	configFile := auth.Group("/configfiles")
	{
		// List all config files - authenticated users see their accessible files
		configFile.GET("", am.Admin(), h.ConfigFile.ListConfigFilesHandler)

		// View specific config file - group member access
		configFile.GET("/:id", am.GroupMember(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.GetConfigFileHandler)

		// Create config file - group manager access (via project_id in body)
		configFile.POST("", am.GroupManager(middleware.FromProjectIDInPayload()), h.ConfigFile.CreateConfigFileHandler)

		// Update config file - group manager access
		configFile.PUT("/:id", am.GroupManager(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.UpdateConfigFileHandler)

		// Delete config file - group manager access
		configFile.DELETE("/:id", am.GroupManager(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.DeleteConfigFileHandler)

		// List config files by project - group member access
		configFile.GET("/project/:project_id", h.ConfigFile.ListConfigFilesByProjectIDHandler)

		// Create instance from config file - group member access
		configFile.POST("/:id/instance", am.GroupMember(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.CreateInstanceHandler)

		// Delete instance - group member access
		configFile.DELETE("/:id/instance", am.GroupMember(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.DestructInstanceHandler)
	}
	auth.POST("/instance/:id", am.GroupMember(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.CreateInstanceHandler)
	auth.DELETE("/instance/:id", am.GroupMember(middleware.FromConfigFileIDParam(repos)), h.ConfigFile.DestructInstanceHandler)
}
