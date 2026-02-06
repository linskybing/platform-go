package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// StoragePermissionHandler handles group storage permission APIs
type StoragePermissionHandler struct {
	permManager *k8s.PermissionManager
}

// NewStoragePermissionHandler creates a new permission handler
func NewStoragePermissionHandler(pm *k8s.PermissionManager) *StoragePermissionHandler {
	return &StoragePermissionHandler{
		permManager: pm,
	}
}

// SetPermission godoc
// @Summary Set user permission for group storage
// @Description Group admin sets permission for a user on a specific PVC
// @Tags Group Storage Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body storage.SetStoragePermissionRequest true "Permission request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/permissions [post]
func (h *StoragePermissionHandler) SetPermission(c *gin.Context) {
	var req storage.SetStoragePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.permManager.SetPermission(c.Request.Context(), &req, adminID); err != nil {
		if err.Error() == "only group admins can set permissions" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to set permission: "+err.Error())
		return
	}

	response.Success(c, nil, "Permission set successfully")
}

// BatchSetPermissions godoc
// @Summary Batch set permissions for multiple users
// @Description Group admin sets permissions for multiple users at once
// @Tags Group Storage Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body storage.BatchSetPermissionsRequest true "Batch permissions request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/permissions/batch [post]
func (h *StoragePermissionHandler) BatchSetPermissions(c *gin.Context) {
	var req storage.BatchSetPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.permManager.BatchSetPermissions(c.Request.Context(), &req, adminID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to batch set permissions: "+err.Error())
		return
	}

	response.Success(c, nil, "Batch permissions set successfully")
}

// GetUserPermission godoc
// @Summary Get user's permission for a specific PVC
// @Description Retrieve permission level for current user on a group PVC
// @Tags Group Storage Permissions
// @Security BearerAuth
// @Produce json
// @Param group_id path int true "Group ID"
// @Param pvc_id path string true "PVC ID"
// @Success 200 {object} response.Response{data=storage.GroupStoragePermission}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/permissions/group/{group_id}/pvc/{pvc_id} [get]
func (h *StoragePermissionHandler) GetUserPermission(c *gin.Context) {
	groupID := c.Param("group_id")
	if groupID == "" {
		response.Error(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	pvcID := c.Param("pvc_id")
	if pvcID == "" {
		response.Error(c, http.StatusBadRequest, "PVC ID is required")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	perm, err := h.permManager.GetUserPermission(c.Request.Context(), userID, groupID, pvcID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get permission: "+err.Error())
		return
	}

	response.Success(c, perm, "Permission retrieved successfully")
}

// SetAccessPolicy godoc
// @Summary Set default access policy for a group PVC
// @Description Group admin sets default permission policy for new members
// @Tags Group Storage Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body storage.SetStorageAccessPolicyRequest true "Access policy request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/policies [post]
func (h *StoragePermissionHandler) SetAccessPolicy(c *gin.Context) {
	var req storage.SetStorageAccessPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.permManager.SetAccessPolicy(c.Request.Context(), &req, adminID); err != nil {
		if err.Error() == "only group admins can set access policies" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to set policy: "+err.Error())
		return
	}

	response.Success(c, nil, "Access policy set successfully")
}
