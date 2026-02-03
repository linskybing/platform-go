package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Service) Ping(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	return s.client.Ping(ctx).Err()
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	if !s.Enabled() {
		return "", ErrCacheMiss
	}
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}
	if err := s.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (s *Service) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if !s.Enabled() {
		return ErrCacheMiss
	}
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal failed: %w", err)
	}
	return nil
}

func (s *Service) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}
	if err := s.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (s *Service) AsyncSetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !s.Enabled() || s.asyncQueue == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}

	req := asyncRequest{ctx: ctx, key: key, payload: payload, ttl: ttl}
	select {
	case s.asyncQueue <- req:
		return nil
	default:
		return fmt.Errorf("async cache queue full")
	}
}

func (s *Service) Invalidate(ctx context.Context, keys ...string) error {
	if !s.Enabled() || len(keys) == 0 {
		return nil
	}
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

func (s *Service) InvalidatePrefix(ctx context.Context, prefix string) error {
	if !s.Enabled() || prefix == "" {
		return nil
	}

	pattern := prefix + "*"
	iter := s.client.Scan(ctx, 0, pattern, 200).Iterator()
	keys := make([]string, 0, 200)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 200 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis delete failed: %w", err)
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan failed: %w", err)
	}
	if len(keys) > 0 {
		if err := s.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("redis delete failed: %w", err)
		}
	}
	return nil
}
