package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/cron"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"gorm.io/gorm"
)

// RegisterRoutes initializes all API routes and middleware.
// Routes are organized by feature into separate files (audit.go, group.go, etc.)
// to maintain modularity and comply with 200-line file limit guidelines.
//
// Organization:
// - Public routes: auth status (no JWT required)
// - Protected routes: grouped under /api with JWT middleware
// - Authorization: role-based access control via auth middleware
func RegisterRoutes(r *gin.Engine, db *gorm.DB, cacheSvc *cache.Service) {
	// Initialize repositories, services, and handlers
	repos := repository.NewRepositories(db)
	services := application.NewWithCache(repos, cacheSvc)
	handlers := handlers.New(services, repos, r)

	// Initialize authorization middleware
	authMiddleware := middleware.NewAuthMiddleware(repos)

	// Start background tasks (cleanup, metrics, etc.)
	cron.StartCleanupTask(services.Audit)

	// Register public routes (no JWT protection)
	registerAuthRoutes(r, handlers)

	// Register protected routes (JWT required)
	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		registerAuditRoutes(auth, handlers, authMiddleware)
		registerConfigFileRoutes(auth, handlers, authMiddleware, repos)
		registerGroupRoutes(auth, handlers, authMiddleware)
		registerUserGroupsRoutes(auth, handlers, authMiddleware)
		registerFormRoutes(auth, handlers, authMiddleware)
		registerImageRoutes(auth, handlers, authMiddleware)
		registerStorageRoutes(auth, handlers, authMiddleware)
		registerProjectRoutes(auth, handlers, authMiddleware)
		registerUserRoutes(auth, handlers, authMiddleware)
		registerK8sRoutes(auth, handlers, authMiddleware)
		registerWebSocketRoutes(auth, services)
	}
}
