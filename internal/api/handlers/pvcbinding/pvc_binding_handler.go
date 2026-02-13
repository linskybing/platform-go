package pvcbinding

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// PVCBindingHandler handles project PVC binding APIs
//
// revive warns on exported type name matching filename; this is a public handler.
//
//nolint:revive
type PVCBindingHandler struct {
	bindingManager *k8s.PVCBindingManager
}

// NewPVCBindingHandler creates a new binding handler
func NewPVCBindingHandler(bm *k8s.PVCBindingManager) *PVCBindingHandler {
	return &PVCBindingHandler{
		bindingManager: bm,
	}
}

// CreateBinding godoc
// @Summary Create PVC binding in project namespace
// @Description Bind group storage to user's project namespace with permission-based access
// @Tags Project PVC Bindings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body storage.CreateProjectPVCBindingRequest true "Binding request"
// @Success 200 {object} response.Response{data=storage.ProjectPVCBindingInfo}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /k8s/pvc-binding [post]
func (h *PVCBindingHandler) CreateBinding(c *gin.Context) {
	var req storage.CreateProjectPVCBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	binding, err := h.bindingManager.CreateProjectPVCBinding(c.Request.Context(), &req, userID)
	if err != nil {
		if errors.Is(err, k8s.ErrPermissionDenied) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create binding: "+err.Error())
		return
	}

	response.Success(c, binding, "PVC binding created successfully")
}

// ListBindings godoc
// @Summary List all PVC bindings for a project
// @Description Get all PVC bindings for a specific project
// @Tags Project PVC Bindings
// @Security BearerAuth
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} response.Response{data=[]storage.ProjectPVCBindingInfo}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /k8s/pvc-binding/project/{project_id} [get]
func (h *PVCBindingHandler) ListBindings(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		response.Error(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	bindings, err := h.bindingManager.ListProjectPVCBindings(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list bindings: "+err.Error())
		return
	}

	response.Success(c, bindings, "PVC bindings retrieved successfully")
}

// DeleteBindingByID godoc
// @Summary Delete PVC binding by binding ID
// @Description Remove a PVC binding by its ID
// @Tags Project PVC Bindings
// @Security BearerAuth
// @Produce json
// @Param binding_id path string true "Binding ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /k8s/pvc-binding/{binding_id} [delete]
func (h *PVCBindingHandler) DeleteBindingByID(c *gin.Context) {
	bindingID := c.Param("binding_id")
	if bindingID == "" {
		response.Error(c, http.StatusBadRequest, "Binding ID is required")
		return
	}

	if err := h.bindingManager.DeleteProjectPVCBindingByID(c.Request.Context(), bindingID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete binding: "+err.Error())
		return
	}

	response.Success(c, nil, "PVC binding deleted successfully")
}

// DeleteBinding godoc
// @Summary Delete PVC binding from project namespace
// @Description Remove a PVC binding and cleanup resources
// @Tags Project PVC Bindings
// @Security BearerAuth
// @Produce json
// @Param project_id path int true "Project ID"
// @Param pvc_name path string true "PVC name in project namespace"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /k8s/pvc-binding/{project_id}/{pvc_name} [delete]
func (h *PVCBindingHandler) DeleteBinding(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		response.Error(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	pvcName := c.Param("pvc_name")
	if pvcName == "" {
		response.Error(c, http.StatusBadRequest, "PVC name is required")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.bindingManager.DeleteProjectPVCBinding(c.Request.Context(), projectID, userID, pvcName); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete binding: "+err.Error())
		return
	}

	response.Success(c, nil, "PVC binding deleted successfully")
}
