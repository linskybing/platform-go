package userstorage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// UserStorageHandler handles HTTP endpoints for admin user storage management
//
// revive warns on exported type name matching filename; this is a public handler.
//
//nolint:revive
type UserStorageHandler struct {
	k8sSvc    *k8s.K8sService
	auditRepo repository.AuditRepo
}

func NewUserStorageHandler(k8sSvc *k8s.K8sService, auditRepo repository.AuditRepo) *UserStorageHandler {
	return &UserStorageHandler{
		k8sSvc:    k8sSvc,
		auditRepo: auditRepo,
	}
}

// CheckStatus godoc
// @Summary Check if user storage exists
// @Tags Admin Storage
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user-storage/{username}/status [get]
func (h *UserStorageHandler) CheckStatus(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "Username is required")
		return
	}

	exists, err := h.k8sSvc.CheckUserStorageExists(c.Request.Context(), username)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check storage status: "+err.Error())
		return
	}

	response.Success(c, gin.H{"exists": exists}, "")
}

// Initialize godoc
// @Summary Initialize user storage hub
// @Tags Admin Storage
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user-storage/{username}/init [post]
func (h *UserStorageHandler) Initialize(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "Username is required")
		return
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Admin ID not found")
		return
	}

	if err := h.k8sSvc.InitializeUserStorageHub(username, adminID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to initialize user storage: "+err.Error())
		return
	}

	// Audit log
	utils.LogAuditWithConsole(c, "create", "user_storage", username, nil, gin.H{"username": username}, "Initialized user storage hub", h.auditRepo)

	response.Success(c, nil, "User storage initialized successfully")
}

// Expand godoc
// @Summary Expand user storage capacity
// @Tags Admin Storage
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param request body map[string]string true "New size (e.g., {\"new_size\": \"20Gi\"})"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user-storage/{username}/expand [put]
func (h *UserStorageHandler) Expand(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "Username is required")
		return
	}

	var req struct {
		NewSize string `json:"new_size" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.k8sSvc.ExpandUserStorageHub(username, req.NewSize); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to expand user storage: "+err.Error())
		return
	}

	// Audit log
	utils.LogAuditWithConsole(c, "update", "user_storage", username, nil, gin.H{"username": username, "new_size": req.NewSize}, "Expanded user storage", h.auditRepo)

	response.Success(c, nil, "User storage expanded successfully")
}

// Delete godoc
// @Summary Delete user storage hub
// @Tags Admin Storage
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user-storage/{username} [delete]
func (h *UserStorageHandler) Delete(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "Username is required")
		return
	}

	if err := h.k8sSvc.DeleteUserStorageHub(c.Request.Context(), username); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete user storage: "+err.Error())
		return
	}

	// Audit log
	utils.LogAuditWithConsole(c, "delete", "user_storage", username, gin.H{"username": username}, nil, "Deleted user storage hub", h.auditRepo)

	response.Success(c, nil, "User storage deleted successfully")
}
