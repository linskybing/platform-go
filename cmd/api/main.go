package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/api/routes"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/cron"
	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/common"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/form"
	"github.com/linskybing/platform-go/internal/domain/gpuusage"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/plugin"
	"github.com/linskybing/platform-go/internal/plugin/builtin"
	jobplugin "github.com/linskybing/platform-go/internal/plugin/builtin/job"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
)

func main() {
	// Load configuration from environment variables and .env file
	if _, err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize JWT signing key
	middleware.Init()

	// Initialize Kubernetes scheme (register API types)
	// config.InitK8sConfig() - removed, handled in LoadConfig

	// Initialize database connection
	db.Init()
	k8s.Init()
	if err := k8s.EnsurePriorityClass(context.Background()); err != nil {
		log.Printf("warning: failed to ensure priority class: %v", err)
	}
	if err := k8s.EnsureConfigFileQueue(context.Background()); err != nil {
		log.Printf("warning: failed to ensure configfile queue: %v", err)
	}

	cacheSvc := cache.NewNoop()
	if config.RedisAddr != "" {
		client, err := cache.NewRedisClient(cache.Config{
			Addr:         config.RedisAddr,
			Username:     config.RedisUsername,
			Password:     config.RedisPassword,
			DB:           config.RedisDB,
			UseTLS:       config.RedisUseTLS,
			PoolSize:     config.RedisPoolSize,
			MinIdleConns: config.RedisMinIdleConns,
			MaxRetries:   config.RedisMaxRetries,
			DialTimeout:  time.Duration(config.RedisDialTimeoutMs) * time.Millisecond,
			ReadTimeout:  time.Duration(config.RedisReadTimeoutMs) * time.Millisecond,
			WriteTimeout: time.Duration(config.RedisWriteTimeoutMs) * time.Millisecond,
			PingTimeout:  time.Duration(config.RedisPingTimeoutMs) * time.Millisecond,
		})
		if err != nil {
			log.Printf("warning: failed to initialize redis client: %v", err)
		} else if client != nil {
			cacheSvc = cache.NewService(
				client,
				cache.WithAsyncQueueSize(config.RedisAsyncQueue),
				cache.WithAsyncWorkers(config.RedisAsyncWorkers),
			)
		}
	}
	defer func() {
		_ = cacheSvc.Close()
	}()

	// Auto migrate database schemas in phases to diagnose FK ordering
	log.Println("AutoMigrating core models: user + project")
	if err := db.DB.AutoMigrate(
		&common.ResourceOwner{},
		&user.User{},
		&project.Project{},
	); err != nil {
		log.Fatalf("Failed to migrate core database models: %v", err)
	}

	log.Println("AutoMigrating group models")
	if err := db.DB.AutoMigrate(
		&group.Group{},
		&group.UserGroup{},
	); err != nil {
		log.Fatalf("Failed to migrate group database models: %v", err)
	}

	log.Println("AutoMigrating remaining models")
	if err := db.DB.AutoMigrate(
		&configfile.ConfigBlob{},
		&configfile.ConfigCommit{},
		&form.Form{},
		&form.FormMessage{},
		&audit.AuditLog{},
		&image.ContainerRepository{},
		&image.ContainerTag{},
		&image.ImageAllowList{},
		&image.ImageRequest{},
		&image.ClusterImageStatus{},
	); err != nil {
		log.Fatalf("Failed to migrate remaining database models: %v", err)
	}

	log.Println("AutoMigrating storage and job models")
	if err := db.DB.AutoMigrate(
		&storage.Storage{},
		&job.PriorityClass{},
		&job.Job{},
		&gpuusage.JobGPUUsageSnapshot{},
		&gpuusage.JobGPUUsageSummary{},
	); err != nil {
		log.Fatalf("Failed to migrate storage and job database models: %v", err)
	}

	// Ensure constraints that require tables to exist
	db.EnsureConstraints()

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Docker cleanup CronJob
	if err := cron.CreateDockerCleanupCronJob(); err != nil {
		log.Printf("Warning: Failed to create Docker cleanup CronJob: %v", err)
		// Don't fail startup if CronJob creation fails
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Register builtin plugins
	builtin.Init()
	plugin.Register(jobplugin.NewJobPlugin())

	// Initialize Plugin Manager
	pluginMgr := plugin.NewManager(db.DB, cacheSvc)
	if err := pluginMgr.Init(); err != nil {
		log.Printf("Warning: failed to initialize plugins: %v", err)
	}
	defer pluginMgr.Shutdown()

	apiGroup := router.Group("/api/v1")
	pluginMgr.RegisterRoutes(apiGroup)

	routes.RegisterRoutes(router, db.DB, cacheSvc)

	port := ":" + config.ServerPort
	log.Printf("Starting API server on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
