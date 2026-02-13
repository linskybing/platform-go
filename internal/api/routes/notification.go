package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerNotificationRoutes registers notification-related routes.
func registerNotificationRoutes(r *gin.RouterGroup, h *handlers.Handlers, _ *middleware.AuthMiddleware) {
	notifications := r.Group("/api/notifications")
	{
		notifications.PUT("/read-all", h.Notification.MarkAllAsRead)
		notifications.DELETE("/clear-all", h.Notification.ClearAll)
		notifications.PUT("/:id/read", h.Notification.MarkAsRead)
	}
}
