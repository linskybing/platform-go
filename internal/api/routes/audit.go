package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerAuditRoutes registers audit log routes.
// All audit routes are protected by JWT middleware with role-based authorization.
//
// Authorization levels:
// - GET audit logs: Admin only (super admin can view all) or Group admin (can view group logs)
// Note: Handler should filter logs based on user's group access
func registerAuditRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	audit := auth.Group("/audit")
	{
		// View audit logs - authenticated users see logs based on their access level
		// Handler filters: Super admin sees all, group admin sees group logs
		audit.GET("", h.Audit.GetAuditLogs)
		audit.GET("/logs", h.Audit.GetAuditLogs)
	}
}
