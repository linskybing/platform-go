package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/types"
)

type JobHandler struct {
	repos    *repository.Repos
	executor executor.Executor
}

func NewJobHandler(repos *repository.Repos, exec executor.Executor) *JobHandler {
	return &JobHandler{
		repos:    repos,
		executor: exec,
	}
}

// ListJobs godoc
// @Summary List jobs for current user
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Success 200 {array} job.Job
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /jobs [get]
func (h *JobHandler) ListJobs(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	userClaims, ok := claims.(*types.Claims)
	if !ok {
		response.Unauthorized(c, "invalid claims")
		return
	}

	jobs, err := h.repos.Job.ListByUser(c.Request.Context(), userClaims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// GetJob godoc
// @Summary Get job by ID
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} job.Job
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /jobs/:id [get]
func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "job ID required")
		return
	}

	job, err := h.repos.Job.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "job not found")
		return
	}

	c.JSON(http.StatusOK, job)
}

// CancelJob godoc
// @Summary Cancel a running or queued job
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} response.MessageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /jobs/:id/cancel [post]
func (h *JobHandler) CancelJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "job ID required")
		return
	}

	if h.executor == nil {
		response.Error(c, http.StatusInternalServerError, "executor not configured")
		return
	}

	err := h.executor.Cancel(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "job cancelled successfully"})
}
