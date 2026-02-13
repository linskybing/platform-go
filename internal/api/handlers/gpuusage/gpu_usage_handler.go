package gpuusage

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appgpuusage "github.com/linskybing/platform-go/internal/application/gpuusage"
	domaingpu "github.com/linskybing/platform-go/internal/domain/gpuusage"
	"github.com/linskybing/platform-go/pkg/response"
	"gorm.io/gorm"
)

type GPUUsageHandler struct {
	svc *appgpuusage.GPUUsageService
}

type GPUUsageListResponse struct {
	Snapshots []domaingpu.JobGPUUsageSnapshot `json:"snapshots"`
	Total     int64                           `json:"total"`
}

func NewGPUUsageHandler(svc *appgpuusage.GPUUsageService) *GPUUsageHandler {
	return &GPUUsageHandler{svc: svc}
}

// GetJobGPUUsage godoc
// @Summary Get job GPU usage snapshots
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} GPUUsageListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /jobs/:id/gpu-usage [get]
func (h *GPUUsageHandler) GetJobGPUUsage(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		response.Error(c, http.StatusBadRequest, "job ID required")
		return
	}

	limit := parseIntQuery(c, "limit", 200)
	offset := parseIntQuery(c, "offset", 0)

	snapshots, total, err := h.svc.ListSnapshots(c.Request.Context(), jobID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, GPUUsageListResponse{Snapshots: snapshots, Total: total})
}

// GetJobGPUSummary godoc
// @Summary Get job GPU usage summary
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} gpuusage.JobGPUUsageSummary
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /jobs/:id/gpu-summary [get]
func (h *GPUUsageHandler) GetJobGPUSummary(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		response.Error(c, http.StatusBadRequest, "job ID required")
		return
	}

	summary, err := h.svc.GetSummary(c.Request.Context(), jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "summary not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, summary)
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return val
}
