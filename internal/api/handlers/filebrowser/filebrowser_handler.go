package filebrowser

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// FileBrowserHandler handles FileBrowser access APIs
//
// revive warns on exported type name matching filename; this is a public handler.
//
//nolint:revive
type FileBrowserHandler struct {
	fbManager *k8s.FileBrowserManager
}

// NewFileBrowserHandler creates a new FileBrowser handler
func NewFileBrowserHandler(fbm *k8s.FileBrowserManager) *FileBrowserHandler {
	return &FileBrowserHandler{
		fbManager: fbm,
	}
}

// GetAccess godoc
// @Summary Get FileBrowser access URL
// @Description Get access to FileBrowser pod based on user's permission (read-only or read-write)
// @Tags FileBrowser
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body storage.FileBrowserAccessRequest true "Access request"
// @Success 200 {object} response.Response{data=storage.FileBrowserAccessResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /k8s/filebrowser/access [post]
func (h *FileBrowserHandler) GetAccess(c *gin.Context) {
	var req storage.FileBrowserAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Override userID with authenticated user
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	req.UserID = userID

	accessResp, err := h.fbManager.GetFileBrowserAccess(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get access: "+err.Error())
		return
	}

	if !accessResp.Allowed {
		response.Error(c, http.StatusForbidden, accessResp.Message)
		return
	}

	response.Success(c, accessResp, "Access granted")
}
