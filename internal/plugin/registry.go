package plugin

import (
	"fmt"
	"log/slog"
	"sync"
)

type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

var GlobalRegistry = &Registry{
	plugins: make(map[string]Plugin),
}

func Register(p Plugin) {
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	if _, exists := GlobalRegistry.plugins[p.Name()]; exists {
		// Just log error or panic? Panic ensures we don't have duplicate plugins at startup.
		panic(fmt.Sprintf("plugin %s already registered", p.Name()))
	}
	GlobalRegistry.plugins[p.Name()] = p
	slog.Info("plugin registered", "name", p.Name(), "version", p.Version())
}

func (r *Registry) InitAll(ctx PluginContext) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, p := range r.plugins {
		slog.Info("initializing plugin", "name", name, "version", p.Version())
		if err := p.Init(ctx); err != nil {
			return fmt.Errorf("failed to init plugin %s: %w", name, err)
		}
	}
	return nil
}

func (r *Registry) GetRoutes() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []Plugin
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}

func (r *Registry) ShutdownAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, p := range r.plugins {
		if err := p.Shutdown(); err != nil {
			slog.Error("failed to shutdown plugin", "name", name, "error", err)
		}
	}
}
