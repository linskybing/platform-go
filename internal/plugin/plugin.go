package plugin

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/events"
	"github.com/linskybing/platform-go/pkg/cache"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

// PluginContext holds dependencies passed to plugins during initialization
type PluginContext struct {
	DB           *gorm.DB
	Cache        *cache.Service
	K8sClient    kubernetes.Interface
	Config       map[string]string
	EventBus     events.EventBus
	HookRegistry *HookRegistry
	Logger       *slog.Logger
}

// Migration defines database changes for a plugin
type Migration struct {
	ID   string
	Up   []string
	Down []string
}

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	Name() string
	Version() string
	Init(ctx PluginContext) error
	RegisterRoutes(router *gin.RouterGroup)
	RegisterMigrations() []Migration
	RegisterEvents(bus events.EventBus)
	Shutdown() error
}
