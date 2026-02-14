package preemptor

import (
	"context"
)

// Strategy defines the interface for different preemption algorithms.
type Strategy interface {
	// Name returns the unique identifier for the strategy.
	Name() string
	// Execute performs the preemption logic and returns victim jobs.
	Execute(ctx context.Context, req ResourceRequirement) (*PreemptionDecision, error)
}

// Registry manages available preemption strategies.
type Registry struct {
	strategies map[string]Strategy
}

// NewRegistry creates a new strategy registry.
func NewRegistry() *Registry {
	return &Registry{strategies: make(map[string]Strategy)}
}

// Register adds a new strategy to the registry.
func (r *Registry) Register(s Strategy) {
	r.strategies[s.Name()] = s
}

// Get retrieves a strategy by name.
func (r *Registry) Get(name string) (Strategy, bool) {
	s, ok := r.strategies[name]
	return s, ok
}
