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
// - GET audit logs: Admin only (super admin can view all)
func registerAuditRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	audit := auth.Group("/audit")
	{
		audit.GET("", am.Admin(), h.Audit.GetAuditLogs)
	}
}
