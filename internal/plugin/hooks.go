package plugin

import (
	"context"
	"sync"
)

type HookType string

const (
	HookBeforeCreate HookType = "BeforeCreate"
	HookAfterCreate  HookType = "AfterCreate"
	HookBeforeUpdate HookType = "BeforeUpdate"
	HookAfterUpdate  HookType = "AfterUpdate"
	HookBeforeDelete HookType = "BeforeDelete"
	HookAfterDelete  HookType = "AfterDelete"
)

type HookFunc func(ctx context.Context, resourceType string, data interface{}) error

type HookRegistry struct {
	hooks map[HookType][]HookFunc
	mu    sync.RWMutex
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookType][]HookFunc),
	}
}

func (r *HookRegistry) Register(hookType HookType, fn HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[hookType] = append(r.hooks[hookType], fn)
}

func (r *HookRegistry) Execute(ctx context.Context, hookType HookType, resourceType string, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if hooks, ok := r.hooks[hookType]; ok {
		for _, fn := range hooks {
			if err := fn(ctx, resourceType, data); err != nil {
				return err
			}
		}
	}
	return nil
}
