package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/linskybing/platform-go/internal/repository"
)

// IDExtractor extracts a group ID from the request context.
// Used by authorization middleware to determine which group a resource belongs to.
type IDExtractor func(*gin.Context, *repository.Repos) string

// FromGroupIDParam extracts group ID directly from URL parameter /:group_id.
func FromGroupIDParam() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) string {
		groupIDStr := c.Param("id")
		if groupIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
			return ""
		}
		return groupIDStr
	}
}

// FromGroupIDParamName extracts group ID from a named URL parameter.
func FromGroupIDParamName(paramName string) IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) string {
		groupIDStr := c.Param(paramName)
		if groupIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
			return ""
		}
		return groupIDStr
	}
}

// FromProjectIDParam extracts group ID by looking up project's group.
// Use when URL has /:id representing a project ID.
func FromProjectIDParam(repos *repository.Repos) IDExtractor {
	return func(c *gin.Context, r *repository.Repos) string {
		projectIDStr := c.Param("id")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
			return ""
		}

		project, err := repos.Project.GetProjectByID(c.Request.Context(), projectIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return ""
		}

		return project.GID
	}
}

// FromProjectIDParamName extracts group ID by looking up project using a named URL parameter.
func FromProjectIDParamName(paramName string) IDExtractor {
	return func(c *gin.Context, r *repository.Repos) string {
		projectIDStr := c.Param(paramName)
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
			return ""
		}

		project, err := r.Project.GetProjectByID(c.Request.Context(), projectIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return ""
		}

		return project.GID
	}
}

// FromConfigCommitIDParam extracts group ID by looking up config commit's project.
// Use when URL has /:id representing a config commit ID.
func FromConfigCommitIDParam(repos *repository.Repos) IDExtractor {
	return func(c *gin.Context, r *repository.Repos) string {
		configIDStr := c.Param("id")
		if configIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config commit ID"})
			return ""
		}

		commit, err := repos.ConfigFile.GetCommit(c.Request.Context(), configIDStr)
		if err != nil || commit == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "config commit not found"})
			return ""
		}

		project, err := repos.Project.GetProjectByID(c.Request.Context(), commit.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return ""
		}

		return project.GID
	}
}

// FromProjectIDInPayload extracts group ID from project_id field in JSON body.
// Use when creating resources that have project_id in the request payload.
func FromProjectIDInPayload() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) string {
		var payload struct {
			ProjectID string `json:"project_id" form:"project_id"`
		}

		if err := bindPayload(c, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return ""
		}

		project, err := repos.Project.GetProjectByID(c.Request.Context(), payload.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return ""
		}

		return project.GID
	}
}

// FromGroupIDInPayload extracts group ID directly from request body.
// Use when creating resources that have group_id in the request payload.
func FromGroupIDInPayload() IDExtractor {
	return func(c *gin.Context, repos *repository.Repos) string {
		var payload struct {
			GroupID string `json:"group_id" form:"group_id"`
		}

		if err := bindPayload(c, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return ""
		}

		if payload.GroupID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
			return ""
		}

		return payload.GroupID
	}
}

func bindPayload(c *gin.Context, payload interface{}) error {
	switch c.ContentType() {
	case binding.MIMEJSON:
		return c.ShouldBindBodyWith(payload, binding.JSON)
	case binding.MIMEXML, binding.MIMEXML2:
		return c.ShouldBindBodyWith(payload, binding.XML)
	default:
		return c.ShouldBind(payload)
	}
}
