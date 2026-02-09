package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	dimage "github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/response"
)

type ImageHandler struct {
	service *application.ImageService
	cache   *cache.Service
	logger  *slog.Logger
}

func NewImageHandler(service *application.ImageService) *ImageHandler {
	return &ImageHandler{
		service: service,
		logger:  slog.Default(),
	}
}

func NewImageHandlerWithCache(service *application.ImageService, cacheSvc *cache.Service, logger *slog.Logger) *ImageHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ImageHandler{
		service: service,
		cache:   cacheSvc,
		logger:  logger,
	}
}

// request/response DTOs live in internal/domain/image

// PullImage triggers async image pull jobs to sync to Harbor.
func (h *ImageHandler) PullImage(c *gin.Context) {
	var payload dimage.PullImageRequestDTO
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}
	requests := make([]dimage.PullRequestDTO, 0)
	if len(payload.Names) > 0 {
		for _, fullImage := range payload.Names {
			name, tag := splitImageName(fullImage)
			requests = append(requests, dimage.PullRequestDTO{Name: name, Tag: tag})
		}
	} else if payload.Name != "" {
		name := payload.Name
		tag := payload.Tag
		if tag == "" {
			name, tag = splitImageName(payload.Name)
		}
		requests = append(requests, dimage.PullRequestDTO{Name: name, Tag: tag})
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "image name is required"})
		return
	}

	jobIDs := make([]string, 0, len(requests))
	for _, req := range requests {
		jobID, err := h.service.PullImageAsync(req.Name, req.Tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
			return
		}
		jobIDs = append(jobIDs, jobID)
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    http.StatusOK,
		Message: "image pull job started",
		Data: gin.H{
			"job_ids": jobIDs,
		},
	})
}

// GetActivePullJobs returns active image pull jobs with caching.
func (h *ImageHandler) GetActivePullJobs(c *gin.Context) {
	const cacheKey = "image:pull:active"
	const cacheTTL = 5 * time.Second // Short TTL for active jobs

	var jobs interface{}

	// Try cache first if available
	if h.cache != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		if data, err := h.cache.Get(ctx, cacheKey); err == nil {
			if err := json.Unmarshal([]byte(data), &jobs); err == nil {
				h.logger.Debug("cache hit for active pull jobs", "key", cacheKey)
				c.JSON(http.StatusOK, response.SuccessResponse{
					Code:    http.StatusOK,
					Message: "success",
					Data:    jobs,
				})
				return
			}
		}
	}

	// Cache miss or no cache - fetch from service
	jobs = h.service.GetActivePullJobs()

	// Update cache asynchronously if available
	if h.cache != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			data, err := json.Marshal(jobs)
			if err != nil {
				h.logger.Error("failed to marshal active jobs for cache", "error", err)
				return
			}

			if err := h.cache.Set(ctx, cacheKey, string(data), cacheTTL); err != nil {
				h.logger.Warn("failed to cache active pull jobs", "key", cacheKey, "error", err)
			}
		}()
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    jobs,
	})
}

// GetFailedPullJobs returns failed image pull jobs with caching.
func (h *ImageHandler) GetFailedPullJobs(c *gin.Context) {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if v := strings.TrimSpace(raw); v != "" {
			if parsed := parsePositiveInt(v); parsed > 0 {
				limit = parsed
			}
		}
	}

	cacheKey := fmt.Sprintf("image:pull:failed:%d", limit)
	const cacheTTL = 30 * time.Second // Longer TTL for failed jobs (less frequent changes)

	var jobs interface{}

	// Try cache first if available
	if h.cache != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		if data, err := h.cache.Get(ctx, cacheKey); err == nil {
			if err := json.Unmarshal([]byte(data), &jobs); err == nil {
				h.logger.Debug("cache hit for failed pull jobs", "key", cacheKey)
				c.JSON(http.StatusOK, response.SuccessResponse{
					Code:    http.StatusOK,
					Message: "success",
					Data:    jobs,
				})
				return
			}
		}
	}

	// Cache miss or no cache - fetch from service
	jobs = h.service.GetFailedPullJobs(limit)

	// Update cache asynchronously if available
	if h.cache != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			data, err := json.Marshal(jobs)
			if err != nil {
				h.logger.Error("failed to marshal failed jobs for cache", "error", err)
				return
			}

			if err := h.cache.Set(ctx, cacheKey, string(data), cacheTTL); err != nil {
				h.logger.Warn("failed to cache failed pull jobs", "key", cacheKey, "error", err)
			}
		}()
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    jobs,
	})
}

func splitImageName(fullImage string) (string, string) {
	lastColon := strings.LastIndex(fullImage, ":")
	if lastColon > 0 {
		return fullImage[:lastColon], fullImage[lastColon+1:]
	}
	return fullImage, "latest"
}

func parsePositiveInt(raw string) int {
	var n int
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
