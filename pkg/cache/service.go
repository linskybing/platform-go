package cache

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	client *redis.Client
	logger *slog.Logger

	sf singleflight.Group

	asyncQueue   chan asyncRequest
	asyncWorkers int
	wg           sync.WaitGroup
	closed       atomic.Bool

	lockTTL  time.Duration
	lockWait time.Duration
}

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		if logger != nil {
			s.logger = logger
		}
	}
}

func WithAsyncQueueSize(size int) Option {
	return func(s *Service) {
		if size > 0 {
			s.asyncQueue = make(chan asyncRequest, size)
		}
	}
}

func WithAsyncWorkers(workers int) Option {
	return func(s *Service) {
		if workers > 0 {
			s.asyncWorkers = workers
		}
	}
}

func WithLockTTL(ttl time.Duration) Option {
	return func(s *Service) {
		if ttl > 0 {
			s.lockTTL = ttl
		}
	}
}

func WithLockWait(wait time.Duration) Option {
	return func(s *Service) {
		if wait > 0 {
			s.lockWait = wait
		}
	}
}

func NewService(client *redis.Client, opts ...Option) *Service {
	svc := &Service{
		client:       client,
		logger:       slog.Default(),
		lockTTL:      5 * time.Second,
		lockWait:     2 * time.Second,
		asyncWorkers: 1,
	}

	if client != nil {
		svc.asyncQueue = make(chan asyncRequest, 256)
	}

	for _, opt := range opts {
		opt(svc)
	}

	if svc.client != nil && svc.asyncQueue != nil {
		if cap(svc.asyncQueue) == 0 {
			svc.asyncQueue = make(chan asyncRequest, 256)
		}
		for i := 0; i < svc.asyncWorkers; i++ {
			svc.wg.Add(1)
			go svc.runAsyncWorker(i)
		}
	}

	return svc
}

func NewNoop() *Service {
	return &Service{}
}

func (s *Service) Enabled() bool {
	return s != nil && s.client != nil
}

func (s *Service) Close() error {
	if s == nil {
		return nil
	}

	if s.closed.CompareAndSwap(false, true) {
		if s.asyncQueue != nil {
			close(s.asyncQueue)
		}
		s.wg.Wait()
	}

	if s.client != nil {
		return s.client.Close()
	}

	return nil
}
