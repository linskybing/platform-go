//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRedisClient(t *testing.T) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6380"
	} else {
		// If REDIS_URL/ADDR format differs, adjust.
		// scripts/run-integration-tests.sh sets REDIS_URL=redis://localhost:6380/0
		// But helpers often need host:port.
		// We can parse or just use the assumption if running locally.
		// Let's try to parse REDIS_URL if present
	}

	// If standard REDIS_URL is present
	url := os.Getenv("REDIS_URL")
	var opts *redis.Options
	var err error

	if url != "" {
		opts, err = redis.ParseURL(url)
		require.NoError(t, err)
	} else {
		opts = &redis.Options{
			Addr: addr,
		}
	}

	client := redis.NewClient(opts)
	require.NoError(t, client.Ping(context.Background()).Err())
	return client
}

func TestRedis_Integration(t *testing.T) {
	client := getRedisClient(t)
	defer client.Close()
	ctx := context.Background()

	t.Run("SetGet", func(t *testing.T) {
		err := client.Set(ctx, "test-key", "test-value", time.Minute).Err()
		require.NoError(t, err)

		val, err := client.Get(ctx, "test-key").Result()
		require.NoError(t, err)
		assert.Equal(t, "test-value", val)
	})

	t.Run("Expiry", func(t *testing.T) {
		err := client.Set(ctx, "test-expire", "val", 100*time.Millisecond).Err()
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
		_, err = client.Get(ctx, "test-expire").Result()
		assert.Equal(t, redis.Nil, err)
	})
}
