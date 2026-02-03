package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestService(t *testing.T) (*Service, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewService(client, WithAsyncQueueSize(16), WithAsyncWorkers(1))

	cleanup := func() {
		_ = svc.Close()
		mr.Close()
	}

	return svc, cleanup
}

func TestAsyncSetJSON(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	ctx := context.Background()
	key := "cache:test:async"
	data := payload{Name: "alice", Age: 30}

	if err := svc.AsyncSetJSON(ctx, key, data, 2*time.Second); err != nil {
		t.Fatalf("AsyncSetJSON failed: %v", err)
	}

	var got payload
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if err := svc.GetJSON(ctx, key, &got); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if got.Name != data.Name || got.Age != data.Age {
		t.Fatalf("unexpected cached value: %+v", got)
	}
}

func TestGetOrFetchJSONSingleflight(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	var calls int32
	fetch := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return "value", nil
	}

	ctx := context.Background()
	key := "cache:test:singleflight"
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var val string
			err := svc.GetOrFetchJSON(ctx, key, time.Minute, &val, func(ctx context.Context) (interface{}, error) {
				return fetch(ctx)
			})
			if err != nil {
				t.Errorf("GetOrFetchJSON failed: %v", err)
				return
			}
			if val != "value" {
				t.Errorf("unexpected value: %s", val)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected single fetch, got %d", calls)
	}
}

func TestInvalidatePrefix(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()

	ctx := context.Background()
	_ = svc.SetJSON(ctx, "cache:test:1", map[string]string{"a": "1"}, time.Minute)
	_ = svc.SetJSON(ctx, "cache:test:2", map[string]string{"a": "2"}, time.Minute)
	_ = svc.SetJSON(ctx, "cache:other:1", map[string]string{"b": "1"}, time.Minute)

	if err := svc.InvalidatePrefix(ctx, "cache:test:"); err != nil {
		t.Fatalf("InvalidatePrefix failed: %v", err)
	}

	var out map[string]string
	if err := svc.GetJSON(ctx, "cache:test:1", &out); err == nil {
		t.Fatalf("expected cache miss after invalidation")
	}
	if err := svc.GetJSON(ctx, "cache:other:1", &out); err != nil {
		t.Fatalf("expected other key to remain, got: %v", err)
	}
}
