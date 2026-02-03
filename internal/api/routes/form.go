package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerFormRoutes registers form management routes.
// All routes are protected by JWT middleware with role-based authorization.
//
// Authorization levels:
// - POST (create): Group manager (can create forms in projects they manage)
// - GET (my forms): Authenticated user (own forms)
// - GET (all forms): Group member (see forms in their groups)
// - PUT (update status): Group manager (manage forms)
// - Messages: Group member (can view/create messages in forms they have access to)
func registerFormRoutes(auth *gin.RouterGroup, h *handlers.Handlers, am *middleware.AuthMiddleware) {
	forms := auth.Group("/forms")
	{
		// Create form - handler should validate project access
		forms.POST("", h.Form.CreateForm)

		// Get my forms - authenticated user sees their own forms
		forms.GET("/my", h.Form.GetMyForms)

		// Get all forms - authenticated users see forms based on group access
		forms.GET("", h.Form.GetAllForms)

		// Update form status - handler should validate access
		forms.PUT("/:id", h.Form.UpdateFormStatus)
		forms.PUT("/:id/status", h.Form.UpdateFormStatus)

		// Create message - handler should validate access to form
		forms.POST("/:id/messages", h.Form.CreateMessage)

		// List messages - handler should validate access to form
		forms.GET("/:id/messages", h.Form.ListMessages)
	}
}
