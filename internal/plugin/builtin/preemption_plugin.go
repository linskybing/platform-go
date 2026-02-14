package builtin

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/events"
	"github.com/linskybing/platform-go/internal/plugin"
	"github.com/linskybing/platform-go/internal/priority/preemptor"
)

// PreemptionPlugin registers the default SQL-based preemption strategy.
type PreemptionPlugin struct{}

// Name returns the plugin identifier.
func (p *PreemptionPlugin) Name() string { return "preemption-plugin" }

// Version returns the plugin version string.
func (p *PreemptionPlugin) Version() string { return "1.0.0" }

// Init registers the default strategy into the PriorityManager.
func (p *PreemptionPlugin) Init(ctx plugin.PluginContext) error {
	strategy := preemptor.NewSQLStrategy(ctx.DB)
	ctx.PriorityManager.RegisterStrategy(strategy)
	return nil
}

// RegisterRoutes is a placeholder for web endpoints.
func (p *PreemptionPlugin) RegisterRoutes(r *gin.RouterGroup) {}

// RegisterMigrations returns empty as migrations are managed globally.
func (p *PreemptionPlugin) RegisterMigrations() []plugin.Migration { return nil }

// RegisterEvents returns empty for this plugin.
func (p *PreemptionPlugin) RegisterEvents(bus events.EventBus) {}

// Shutdown performs cleanup.
func (p *PreemptionPlugin) Shutdown() error { return nil }
