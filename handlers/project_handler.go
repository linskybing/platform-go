package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/response"
	"github.com/linskybing/platform-go/services"
)

// GetProjects godoc
// @Summary List all projects
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Project
// @Failure 500 {object} response.ErrorResponse
// @Router /projects [get]
func GetProjects(c *gin.Context) {
	projects, err := services.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

// GetProjectByID godoc
// @Summary Get project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} response.ErrorResponse "Invalid project id"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Router /projects/{id} [get]
func GetProjectByID(c *gin.Context) {
	id := c.Param("id")
	project, err := services.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

// CreateProject godoc
// @Summary Create a new project
// @Tags projects
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param project_name formData string true "Project name"
// @Param description formData string false "Description"
// @Param g_id formData int true "Group ID"
// @Success 201 {object} models.Project
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /projects [post]
func CreateProject(c *gin.Context) {
	var input dto.CreateProjectDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	project, err := services.CreateProject(c, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// UpdateProject godoc
// @Summary Update project by ID
// @Tags projects
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Project ID"
// @Param project_name formData string false "Project name"
// @Param description formData string false "Description"
// @Param g_id formData int false "Group ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /projects/{id} [put]
func UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var input dto.UpdateProjectDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	project, err := services.UpdateProject(c, id, input)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary Delete project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.MessageResponse "Project deleted"
// @Failure 400 {object} response.ErrorResponse "Invalid project id"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /projects/{id} [delete]
func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	err := services.DeleteProject(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.MessageResponse{Message: "project deleted"})
}
