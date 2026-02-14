package job

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/types"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type JobHandler struct {
	repos    *repository.Repos
	executor executor.Executor
	config   *configfile.ConfigFileService
}

func NewJobHandler(repos *repository.Repos, exec executor.Executor, configSvc *configfile.ConfigFileService) *JobHandler {
	return &JobHandler{
		repos:    repos,
		executor: exec,
		config:   configSvc,
	}
}

type SubmitJobRequest struct {
	ConfigCommitID string `json:"config_commit_id"`
	ProjectID      string `json:"project_id"`
	SubmitType     string `json:"submit_type"`
	QueueName      string `json:"queue_name"`
	Priority       int32  `json:"priority"`
}

// ListTemplates godoc
// @Summary List available job templates (config commits)
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Success 200 {object} response.StandardResponse{data=[]configfile.ConfigCommit}
// @Failure 401 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /jobs/templates [get]
func (h *JobHandler) ListTemplates(c *gin.Context) {
	projectID := c.Query("project_id")

	var err error
	if projectID != "" {
		templates, err := h.config.ListConfigFilesByProjectID(projectID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Success(c, templates, "Job templates retrieved successfully")
		return
	}

	templates, err := h.config.ListConfigFiles()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, templates, "Job templates retrieved successfully")
}

// ListJobs godoc
// @Summary List jobs for current user
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]job.Job}
// @Failure 401 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
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

	if h.executor != nil {
		for i := range jobs {
			status, err := h.executor.Status(c.Request.Context(), jobs[i].ID)
			if err == nil && status != "" {
				jobs[i].Status = string(status)
			}
		}
	}

	response.Success(c, jobs, "Jobs retrieved successfully")
}

// GetJob godoc
// @Summary Get job by ID
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} response.StandardResponse{data=job.Job}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
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

	if h.executor != nil {
		status, err := h.executor.Status(c.Request.Context(), job.ID)
		if err == nil && status != "" {
			job.Status = string(status)
		}
	}

	response.Success(c, job, "Job retrieved successfully")
}

// CancelJob godoc
// @Summary Cancel a running or queued job
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} response.StandardResponse{data=nil}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
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

	response.Success(c, nil, "Job cancelled successfully")
}

// SubmitJob godoc
// @Summary Submit a job via scheduler
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body SubmitJobRequest true "Submit request"
// @Success 200 {object} response.StandardResponse{data=map[string]string}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /jobs/submit [post]
func (h *JobHandler) SubmitJob(c *gin.Context) {
	if h.config == nil {
		response.Error(c, http.StatusInternalServerError, "config file service not configured")
		return
	}

	var req SubmitJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if req.ConfigCommitID == "" {
		response.Error(c, http.StatusBadRequest, "config_commit_id required")
		return
	}

	submitType := strings.ToLower(strings.TrimSpace(req.SubmitType))
	if submitType == "" {
		submitType = string(executor.SubmitTypeJob)
	}
	if submitType != string(executor.SubmitTypeJob) && submitType != string(executor.SubmitTypeWorkflow) {
		response.Error(c, http.StatusBadRequest, "submit_type must be job or workflow")
		return
	}
	if submitType == string(executor.SubmitTypeWorkflow) && config.ExecutorMode == "scheduler" {
		response.Error(c, http.StatusBadRequest, "workflow submission is not supported in scheduler mode")
		return
	}

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

	commit, err := h.config.GetConfigFile(req.ConfigCommitID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config_commit_id")
		return
	}
	if req.ProjectID != "" && req.ProjectID != commit.ProjectID {
		response.Error(c, http.StatusBadRequest, "project_id does not match config commit")
		return
	}
	proj, err := h.repos.Project.GetProjectByID(c.Request.Context(), commit.ProjectID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "project not found")
		return
	}
	if allowed, err := proj.IsTimeAllowed(time.Now()); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	} else if !allowed {
		response.Error(c, http.StatusBadRequest, "project is outside allowed schedule")
		return
	}
	if err := enforceUserJobLimits(c.Request.Context(), h.repos, proj, userClaims.UserID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	jobID, err := gonanoid.New()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate job id")
		return
	}

	ctx := c.Request.Context()
	ctx = configfile.WithJobID(ctx, jobID)
	ctx = configfile.WithSubmitType(ctx, submitType)
	queueName := strings.TrimSpace(req.QueueName)
	if queueName == "" {
		queueName = config.DefaultQueueName
	}
	if queueName != "" {
		ctx = configfile.WithQueueName(ctx, queueName)
	}
	if req.Priority != 0 {
		ctx = configfile.WithPriority(ctx, req.Priority)
	}

	if err := h.config.CreateInstance(ctx, req.ConfigCommitID, userClaims); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"job_id": jobID}, "Job submitted successfully")
}

func enforceUserJobLimits(ctx context.Context, repos *repository.Repos, proj *project.Project, userID string) error {
	if repos == nil || repos.Job == nil {
		return nil
	}
	if proj == nil {
		return nil
	}
	if proj.MaxConcurrentJobsPerUser > 0 {
		count, err := repos.Job.CountByUserProjectAndStatuses(ctx, userID, proj.PID, []string{string(executor.JobStatusRunning)})
		if err != nil {
			return err
		}
		if count >= int64(proj.MaxConcurrentJobsPerUser) {
			return fmt.Errorf("max concurrent jobs exceeded")
		}
	}
	if proj.MaxQueuedJobsPerUser > 0 {
		count, err := repos.Job.CountByUserProjectAndStatuses(ctx, userID, proj.PID, []string{string(executor.JobStatusQueued), string(executor.JobStatusSubmitted)})
		if err != nil {
			return err
		}
		if count >= int64(proj.MaxQueuedJobsPerUser) {
			return fmt.Errorf("max queued jobs exceeded")
		}
	}
	return nil
}
