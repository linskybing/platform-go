package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const lockPrefix = "lock:"

func (s *Service) acquireLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	if !s.Enabled() {
		return "", false, nil
	}
	value := uuid.NewString()
	ok, err := s.client.SetNX(ctx, lockPrefix+key, value, ttl).Result()
	if err != nil {
		return "", false, fmt.Errorf("lock acquire failed: %w", err)
	}
	return value, ok, nil
}

func (s *Service) releaseLock(ctx context.Context, key, value string) error {
	if !s.Enabled() {
		return nil
	}

	script := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

	if err := s.client.Eval(ctx, script, []string{lockPrefix + key}, value).Err(); err != nil {
		return fmt.Errorf("lock release failed: %w", err)
	}
	return nil
}
