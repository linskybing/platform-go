package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/pkg/response"
)

// NotificationHandler handles notification API endpoints.
// Notifications are currently managed per-session on the frontend via WebSocket.
// These endpoints provide server-side acknowledgement stubs that will be
// backed by a persistent store in a future iteration.
type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// MarkAsRead marks a single notification as read.
// PUT /api/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "Notification ID required")
		return
	}

	// Stub: acknowledge receipt. A future implementation should update a
	// notifications table in the DB.
	response.Success(c, nil, "Notification marked as read")
}

// MarkAllAsRead marks all notifications as read for the current user.
// PUT /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	response.Success(c, nil, "All notifications marked as read")
}

// ClearAll deletes all notifications for the current user.
// DELETE /api/notifications/clear-all
func (h *NotificationHandler) ClearAll(c *gin.Context) {
	response.Success(c, nil, "All notifications cleared")
}
