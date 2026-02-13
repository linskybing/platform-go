package events

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type HandlerFunc func(ctx context.Context, event Event) error

type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType EventType, handler HandlerFunc)
	Unsubscribe(eventType EventType, handler HandlerFunc)
}

type MemoryEventBus struct {
	handlers map[EventType][]HandlerFunc
	lock     sync.RWMutex
}

func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		handlers: make(map[EventType][]HandlerFunc),
	}
}

func (b *MemoryEventBus) Publish(ctx context.Context, event Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	b.lock.RLock()
	handlers, ok := b.handlers[event.Type]
	b.lock.RUnlock()

	if !ok {
		return nil
	}

	// For synchronous hooks/events within the same process, we might want to wait.
	// However, the plan specified "AfterCreate hook used for notification/webhook (non-blocking, via event bus)".
	// So we should run these asynchronously to not block the API request.
	for _, h := range handlers {
		go func(handler HandlerFunc) {
			// Create a background context to ensure execution continues even if request context is cancelled
			ctx := context.WithoutCancel(ctx)
			if err := handler(ctx, event); err != nil {
				slog.Error("failed to handle event",
					"type", event.Type,
					"event_id", event.ID,
					"error", err)
			}
		}(h)
	}

	return nil
}

func (b *MemoryEventBus) Subscribe(eventType EventType, handler HandlerFunc) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *MemoryEventBus) Unsubscribe(eventType EventType, handler HandlerFunc) {
	// Simple unsubscribe not implemented for now
}
