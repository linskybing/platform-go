package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/cron"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

// RegisterRoutes orchestrates the registration of all API routes.
// It initializes dependencies (repositories, services, handlers) and registers all route groups:
//   - Authentication routes (register, login, logout)
//   - WebSocket routes (real-time communication)
//   - Image management routes
//   - Configuration file routes
//   - User management routes
//   - Group and project routes
//   - Kubernetes storage routes
//   - Form management routes
//
// The routes are protected by JWT middleware and use role-based access control (RBAC)
// through custom authorization middleware.
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize repositories
	repos_instance := repository.NewRepositories(db)

	// Initialize services
	services_instance := application.New(repos_instance)

	// Initialize handlers
	handlers_instance := handlers.New(services_instance, repos_instance, r)

	// Initialize auth middleware for RBAC
	authMiddleware := middleware.NewAuth(repos_instance)

	// Start background tasks (cleanup, etc.)
	cron.StartCleanupTask(services_instance.Audit)

	// Register authentication routes (no JWT protection)
	registerAuthRoutes(r, handlers_instance, authMiddleware)

	// Create JWT-protected route group
	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		// Register WebSocket routes
		registerWebSocketRoutes(r, auth, handlers_instance, services_instance)

		// Register feature-specific routes
		registerImageRoutes(auth, handlers_instance, authMiddleware)

		registerConfigFileRoutes(auth, handlers_instance, authMiddleware, repos_instance)

		registerUserRoutes(auth, handlers_instance, authMiddleware)

		registerGroupRoutes(auth, handlers_instance, authMiddleware, repos_instance)

		// Kubernetes routes
		k8sRoutes := auth.Group("/k8s")
		{
			// Register storage routes in K8s group
			registerStorageRoutes(k8sRoutes, handlers_instance, authMiddleware, repos_instance)
		}

		registerFormRoutes(auth, handlers_instance, authMiddleware)
	}
}
