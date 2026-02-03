package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
)

// registerAuthRoutes registers authentication-related routes.
// Auth routes do not require JWT protection.
func registerAuthRoutes(r *gin.Engine) {
	// Public authentication endpoints
	r.GET("/auth/status", handlers.AuthStatusHandler)
	// TODO: Add more auth routes (login, register, logout) when handlers are implemented
}
