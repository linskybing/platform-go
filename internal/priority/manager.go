package priority

import (
	"context"
	"fmt"
	"sync"

	"github.com/linskybing/platform-go/internal/priority/preemptor"
)

// Manager coordinates job priority and preemption workflows.
type Manager struct {
	mu         sync.RWMutex
	strategies *preemptor.Registry
	active     string
}

// NewManager initializes the priority manager with a registry.
func NewManager() *Manager {
	return &Manager{
		strategies: preemptor.NewRegistry(),
		active:     "sql-preemption",
	}
}

// RegisterStrategy adds a new preemption strategy to the manager.
func (m *Manager) RegisterStrategy(s preemptor.Strategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategies.Register(s)
}

// SetActiveStrategy changes the currently used preemption algorithm.
func (m *Manager) SetActiveStrategy(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.strategies.Get(name); !ok {
		return fmt.Errorf("strategy %s not found", name)
	}
	m.active = name
	return nil
}

// Preempt find victims based on the active strategy.
func (m *Manager) Preempt(ctx context.Context, req preemptor.ResourceRequirement) (*preemptor.PreemptionDecision, error) {
	m.mu.RLock()
	strategyName := m.active
	m.mu.RUnlock()

	s, ok := m.strategies.Get(strategyName)
	if !ok {
		return nil, fmt.Errorf("active strategy %s unavailable", strategyName)
	}

	return s.Execute(ctx, req)
}
