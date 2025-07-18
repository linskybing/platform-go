package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/response"
	"github.com/linskybing/platform-go/services"
)

// GetAuditLogs godoc
// @Summary      Query audit logs
// @Tags         audit
// @Security     BearerAuth
// @Produce      json
// @Param        user_id       query     int    false  "User ID"
// @Param        resource_type query     string false  "Resource Type"
// @Param        action        query     string false  "Action"
// @Param        start_time    query     string false  "Start Time (RFC3339)"
// @Param        end_time      query     string false  "End Time (RFC3339)"
// @Param        limit         query     int    false  "Limit"
// @Param        offset        query     int    false  "Offset"
// @Success      200 {array} models.AuditLog
// @Failure      500 {object} response.ErrorResponse
// @Router       /audit/logs [get]
func GetAuditLogs(c *gin.Context) {
	var params repositories.AuditQueryParams

	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			uidUint := uint(uid)
			params.UserID = &uidUint
		}
	}

	if rt := c.Query("resource_type"); rt != "" {
		params.ResourceType = &rt
	}

	if act := c.Query("action"); act != "" {
		params.Action = &act
	}

	if start := c.Query("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			params.StartTime = &t
		}
	}
	if end := c.Query("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			params.EndTime = &t
		}
	}

	// Optional: pagination
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	params.Limit = limit
	params.Offset = offset

	logs, err := services.QueryAuditLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
