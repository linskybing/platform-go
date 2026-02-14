package plugin

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/events"
	"github.com/linskybing/platform-go/internal/priority"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
	"gorm.io/gorm"
	k8sclient "k8s.io/client-go/kubernetes"
)

// Manager handles the lifecycle of plugins
type Manager struct {
	ctx      PluginContext
	registry *Registry
}

func NewManager(db *gorm.DB, cacheSvc *cache.Service) *Manager {
	var client k8sclient.Interface
	if c, ok := k8s.Clientset.(k8sclient.Interface); ok {
		client = c
	}

	priorityMgr := priority.NewManager()

	pCtx := PluginContext{
		DB:              db,
		Cache:           cacheSvc,
		K8sClient:       client,
		Config:          make(map[string]string),
		EventBus:        events.NewMemoryEventBus(),
		HookRegistry:    NewHookRegistry(),
		PriorityManager: priorityMgr,
		Logger:          slog.Default(),
	}

	return &Manager{
		ctx:      pCtx,
		registry: GlobalRegistry,
	}
}

func (m *Manager) Init() error {
	return m.registry.InitAll(m.ctx)
}

func (m *Manager) RegisterRoutes(r *gin.RouterGroup) {
	plugins := m.registry.GetRoutes()
	for _, p := range plugins {
		p.RegisterRoutes(r)
	}
}

func (m *Manager) Shutdown() {
	m.registry.ShutdownAll()
}
