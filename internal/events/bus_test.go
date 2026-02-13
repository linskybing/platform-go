package events

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryEventBus(t *testing.T) {
	bus := NewMemoryEventBus()
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.handlers)
}

func TestPublish_NoSubscribers(t *testing.T) {
	bus := NewMemoryEventBus()
	event := Event{
		Type: JobCreated,
	}
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
}

func TestPublish_WithSubscriber(t *testing.T) {
	bus := NewMemoryEventBus()
	wg := sync.WaitGroup{}
	wg.Add(1)

	var receivedEvent Event
	handler := func(ctx context.Context, e Event) error {
		receivedEvent = e
		wg.Done()
		return nil
	}

	bus.Subscribe(JobCreated, handler)

	event := Event{
		Type:    JobCreated,
		Payload: "test-payload",
	}

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)

	wg.Wait()
	assert.Equal(t, JobCreated, receivedEvent.Type)
	assert.Equal(t, "test-payload", receivedEvent.Payload)
	assert.NotEmpty(t, receivedEvent.ID)
	assert.False(t, receivedEvent.Timestamp.IsZero())
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	bus := NewMemoryEventBus()
	wg := sync.WaitGroup{}
	wg.Add(2)

	handler1 := func(ctx context.Context, e Event) error {
		wg.Done()
		return nil
	}
	handler2 := func(ctx context.Context, e Event) error {
		wg.Done()
		return nil
	}

	bus.Subscribe(JobCreated, handler1)
	bus.Subscribe(JobCreated, handler2)

	err := bus.Publish(context.Background(), Event{Type: JobCreated})
	assert.NoError(t, err)

	wg.Wait()
}

func TestSubscribe_SameEventMultipleTimes(t *testing.T) {
	bus := NewMemoryEventBus()
	wg := sync.WaitGroup{}
	wg.Add(2) // Expect 2 calls

	count := 0
	mu := sync.Mutex{}
	handler := func(ctx context.Context, e Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
		return nil
	}

	// Subscribe twice with same handler logic (different function pointers due to closure, but testing multiple subscribers)
	// Note: MemoryEventBus doesn't dedup handlers
	bus.Subscribe(JobCreated, handler)
	bus.Subscribe(JobCreated, handler)

	err := bus.Publish(context.Background(), Event{Type: JobCreated})
	assert.NoError(t, err)

	wg.Wait()
	mu.Lock()
	assert.Equal(t, 2, count)
	mu.Unlock()
}

func TestUnsubscribe_NoOp(t *testing.T) {
	bus := NewMemoryEventBus()
	handler := func(ctx context.Context, e Event) error { return nil }

	// Should not panic
	bus.Unsubscribe(JobCreated, handler)
}

func TestPublish_HandlerError_LogsButDoesNotFail(t *testing.T) {
	bus := NewMemoryEventBus()
	wg := sync.WaitGroup{}
	wg.Add(1)

	handler := func(ctx context.Context, e Event) error {
		wg.Done()
		return errors.New("handler error")
	}

	bus.Subscribe(JobCreated, handler)

	err := bus.Publish(context.Background(), Event{Type: JobCreated})
	assert.NoError(t, err) // Publish should verify succeed even if handler fails

	wg.Wait()
}

func BenchmarkPublish(b *testing.B) {
	bus := NewMemoryEventBus()
	handler := func(ctx context.Context, e Event) error {
		return nil
	}

	// Setup subscribers
	for i := 0; i < 10; i++ {
		bus.Subscribe(JobCreated, handler)
	}

	ctx := context.Background()
	event := Event{Type: JobCreated}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.Publish(ctx, event)
	}
}
