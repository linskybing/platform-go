package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

// registerAuthRoutes registers authentication-related routes.
// Auth routes do not require JWT protection (except logout/status if needed).
func registerAuthRoutes(r *gin.Engine, h *handlers.Handlers) {
	// Public authentication endpoints
	r.POST("/login", h.User.Login)
	r.POST("/register", h.User.Register)

	// Auth status (requires auth, but placed here for logical grouping, check middleware usage)
	// r.GET("/auth/status", middleware.JWTAuthMiddleware(), handlers.AuthStatusHandler) // This was using standalone handler

	// TODO: AuthStatusHandler should ideally be part of UserHandler or similar
	r.GET("/auth/status", middleware.JWTAuthMiddleware(), handlers.AuthStatusHandler)
}
