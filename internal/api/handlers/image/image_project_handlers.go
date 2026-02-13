package image

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dimage "github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// ListAllowed returns allowed images for a project.
func (h *ImageHandler) ListAllowed(c *gin.Context) {
	pid, err := utils.ParseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project id"})
		return
	}

	images, err := h.service.ListAllowedImages(&pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

// AddProjectImage creates an allow-list request for a project-scoped image.
func (h *ImageHandler) AddProjectImage(c *gin.Context) {
	pid, err := utils.ParseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project id"})
		return
	}

	var req dimage.AddProjectImageDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.service.AddProjectImage(userID, pid, req.ImageName, req.Tag); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "image request submitted"})
}

// RemoveProjectImage disables an allow-list rule (project image) by id.
func (h *ImageHandler) RemoveProjectImage(c *gin.Context) {
	imageID, err := utils.ParseIDParam(c, "image_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid image id"})
		return
	}

	if err := h.service.DisableAllowListRule(imageID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRequestsByProject lists image requests for a specific project.
func (h *ImageHandler) ListRequestsByProject(c *gin.Context) {
	pid, err := utils.ParseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project id"})
		return
	}

	status := c.Query("status")

	requests, err := h.service.ListRequests(&pid, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}
