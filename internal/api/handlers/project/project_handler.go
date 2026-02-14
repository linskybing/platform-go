package project

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
)

type ProjectHandler struct {
	svc *application.ProjectService
}

func NewProjectHandler(svc *application.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// GetProjects godoc
// @Summary List projects for current user
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]project.Project}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /projects [get]
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	// Get project views for this user
	projectViews, err := h.svc.ListProjects()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return empty array if no projects
	if len(projectViews) == 0 {
		response.Success(c, []project.Project{}, "No projects found")
		return
	}

	// Get full project details for each project ID
	result := make([]project.Project, 0, len(projectViews))
	seenPIDs := make(map[string]bool)

	for _, pv := range projectViews {
		// Avoid duplicates (user might be in multiple groups for same project)
		if seenPIDs[pv.PID] {
			continue
		}
		seenPIDs[pv.PID] = true

		// Get full project details from repository
		fullProject, err := h.svc.GetProject(pv.PID)
		if err == nil {
			result = append(result, *fullProject)
		}
	}

	response.Success(c, result, "Projects retrieved successfully")
}

// GetProjectsByUser godoc
// @Summary List projects by user
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.StandardResponse{data=map[string]project.GroupProjects}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /projects/by-user [get]
func (h *ProjectHandler) GetProjectsByUser(c *gin.Context) {
	id, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	records, err := h.svc.GetProjectsByUser(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	grouped := h.svc.GroupProjectsByGID(records)

	// Return empty object if no projects
	if grouped == nil {
		grouped = make(map[string]map[string]interface{})
	}

	response.Success(c, grouped, "Projects by user retrieved successfully")
}

// GetProjectByID godoc
// @Summary Get project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Project ID"
// @Success 200 {object} response.StandardResponse{data=project.Project}
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid project id"
// @Failure 404 {object} response.StandardResponse{data=nil} "Project not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID")
		return
	}
	project, err := h.svc.GetProject(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Project not found")
		return
	}
	response.Success(c, project, "Project retrieved successfully")
}

// CreateProject godoc
// @Summary Create a new project
// @Tags projects
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param project_name formData string true "Project name"
// @Param description formData string false "Description"
// @Param g_id formData uint true "Group ID"
// @Success 201 {object} response.StandardResponse{data=project.Project}
// @Failure 400 {object} response.StandardResponse{data=nil} "Bad request"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var input project.CreateProjectDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Only super admin can set GPU quota and access
	claimsVal, _ := c.Get("claims")
	claims := claimsVal.(*types.Claims)
	if !claims.IsAdmin {
		input.GPUQuota = nil
		// input.GPUAccess = nil
	}

	project, err := h.svc.CreateProject(c, input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, project, "Project created successfully")
}

// UpdateProject godoc
// @Summary Update project by ID
// @Tags projects
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path uint true "Project ID"
// @Param project_name formData string false "Project name"
// @Param description formData string false "Description"
// @Param g_id formData uint false "Group ID"
// @Success 200 {object} response.StandardResponse{data=project.Project}
// @Failure 400 {object} response.StandardResponse{data=nil} "Bad request"
// @Failure 404 {object} response.StandardResponse{data=nil} "Project not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID")
		return
	}
	var input project.UpdateProjectDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Only super admin can set GPU quota and access
	claimsVal, _ := c.Get("claims")
	claims := claimsVal.(*types.Claims)
	if !claims.IsAdmin {
		input.GPUQuota = nil
		// input.GPUAccess = nil
	}

	project, err := h.svc.UpdateProject(c, id, input)
	if err != nil {
		if errors.Is(err, application.ErrProjectNotFound) {
			response.Error(c, http.StatusNotFound, "Project not found")
		} else {
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, project, "Project updated successfully")
}

// DeleteProject godoc
// @Summary Delete project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Project ID"
// @Success 200 {object} response.StandardResponse{data=nil} "Project deleted"
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid project id"
// @Failure 404 {object} response.StandardResponse{data=nil} "Project not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = h.svc.DeleteProject(c, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil, "Project deleted successfully")
}
