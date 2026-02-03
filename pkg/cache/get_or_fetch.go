package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Service) GetOrFetchJSON(ctx context.Context, key string, ttl time.Duration, dest interface{}, fetch func(context.Context) (interface{}, error)) error {
	if dest == nil {
		return fmt.Errorf("dest cannot be nil")
	}

	if !s.Enabled() {
		val, err := fetch(ctx)
		if err != nil {
			return err
		}
		return assignValue(dest, val)
	}

	if data, err := s.getBytes(ctx, key); err == nil {
		return json.Unmarshal(data, dest)
	}

	res, err, _ := s.sf.Do(key, func() (interface{}, error) {
		if data, err := s.getBytes(ctx, key); err == nil {
			return data, nil
		}

		return s.fetchWithLock(ctx, key, ttl, fetch)
	})
	if err != nil {
		return err
	}

	data, ok := res.([]byte)
	if !ok {
		return ErrCacheMiss
	}
	return json.Unmarshal(data, dest)
}

func (s *Service) fetchWithLock(ctx context.Context, key string, ttl time.Duration, fetch func(context.Context) (interface{}, error)) ([]byte, error) {
	lockValue, locked, err := s.acquireLock(ctx, key, s.lockTTL)
	if err != nil {
		return nil, err
	}

	if locked {
		defer func() {
			_ = s.releaseLock(ctx, key, lockValue)
		}()

		val, err := fetch(ctx)
		if err != nil {
			return nil, err
		}

		payload, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("cache marshal failed: %w", err)
		}

		if err := s.client.Set(ctx, key, payload, ttl).Err(); err != nil {
			s.logger.Warn("cache set failed after fetch",
				"key", key,
				"error", err)
		}
		return payload, nil
	}

	if s.lockWait > 0 {
		if data, err := s.waitForBytes(ctx, key, s.lockWait); err == nil {
			return data, nil
		}
	}

	val, err := fetch(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("cache marshal failed: %w", err)
	}
	_ = s.AsyncSetJSON(ctx, key, val, ttl)
	return payload, nil
}

func (s *Service) waitForBytes(ctx context.Context, key string, wait time.Duration) ([]byte, error) {
	deadline := time.Now().Add(wait)

	for time.Now().Before(deadline) {
		if data, err := s.getBytes(ctx, key); err == nil {
			return data, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	return nil, ErrCacheMiss
}

func (s *Service) getBytes(ctx context.Context, key string) ([]byte, error) {
	if !s.Enabled() {
		return nil, ErrCacheMiss
	}
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return data, nil
}

func assignValue(dest interface{}, val interface{}) error {
	payload, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}
	if err := json.Unmarshal(payload, dest); err != nil {
		return fmt.Errorf("cache unmarshal failed: %w", err)
	}
	return nil
}
