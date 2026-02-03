package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/repository"
)

// IDExtractor extracts a group ID from the request context.
// Used by authorization middleware to determine which group a resource belongs to.
type IDExtractor func(*gin.Context, *repository.Repos) uint

// FromGroupIDParam extracts group ID directly from URL parameter /:group_id.
func FromGroupIDParam() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) uint {
		groupIDStr := c.Param("id")
		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
			return 0
		}
		return uint(groupID)
	}
}

// FromProjectIDParam extracts group ID by looking up project's group.
// Use when URL has /:id representing a project ID.
func FromProjectIDParam(repos *repository.Repos) IDExtractor {
	return func(c *gin.Context, r *repository.Repos) uint {
		projectIDStr := c.Param("id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
			return 0
		}

		project, err := repos.Project.GetProjectByID(uint(projectID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return 0
		}

		return project.GID
	}
}

// FromConfigFileIDParam extracts group ID by looking up config file's project's group.
// Use when URL has /:id representing a config file ID.
func FromConfigFileIDParam(repos *repository.Repos) IDExtractor {
	return func(c *gin.Context, r *repository.Repos) uint {
		configIDStr := c.Param("id")
		configID, err := strconv.ParseUint(configIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config file ID"})
			return 0
		}

		configFile, err := repos.ConfigFile.GetConfigFileByID(uint(configID))
		if err != nil || configFile == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "config file not found"})
			return 0
		}

		project, err := repos.Project.GetProjectByID(configFile.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return 0
		}

		return project.GID
	}
}

// FromProjectIDInPayload extracts group ID from project_id field in JSON body.
// Use when creating resources that have project_id in the request payload.
func FromProjectIDInPayload() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) uint {
		var payload struct {
			ProjectID uint `json:"project_id" form:"project_id"`
		}

		if err := c.ShouldBind(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return 0
		}

		project, err := repos.Project.GetProjectByID(payload.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return 0
		}

		return project.GID
	}
}

// FromGroupIDInPayload extracts group ID directly from request body.
// Use when creating resources that have group_id in the request payload.
func FromGroupIDInPayload() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) uint {
		var payload struct {
			GroupID uint `json:"group_id" form:"group_id"`
		}

		if err := c.ShouldBind(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return 0
		}

		if payload.GroupID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
			return 0
		}

		return payload.GroupID
	}
}
