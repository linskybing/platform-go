package routes

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/cron"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
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
	cron.StartClusterResourceCollector(services.Cluster)
	cron.StartGPUUsageCollector(services.GPUUsage)
	if config.FlashSchedEnabled || config.ExecutorMode == "scheduler" {
		flashJobClient := k8s.NewFlashJobClient(k8s.DynamicClient)
		executor.StartFlashJobReconciler(context.Background(), repos, flashJobClient)
	}

	// Register public routes (no JWT protection)
	registerAuthRoutes(r, handlers)

	// Register protected routes (JWT required)
	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		registerClusterRoutes(auth, handlers, authMiddleware)
		registerAuditRoutes(auth, handlers, authMiddleware)
		registerConfigFileRoutes(auth, handlers, authMiddleware, repos)
		registerGroupRoutes(auth, handlers, authMiddleware)
		registerUserGroupsRoutes(auth, handlers, authMiddleware)
		registerFormRoutes(auth, handlers, authMiddleware)
		registerImageRoutes(auth, handlers, authMiddleware)
		// Job routes are now handled by the job plugin
		registerStorageRoutes(auth, handlers, authMiddleware)
		registerAdminStorageRoutes(auth, handlers, authMiddleware)
		registerProjectRoutes(auth, handlers, authMiddleware, repos)
		registerUserRoutes(auth, handlers, authMiddleware)
		registerK8sRoutes(auth, handlers, authMiddleware)
		registerNotificationRoutes(auth, handlers, authMiddleware)
		registerWebSocketRoutes(auth, services)
	}
}
