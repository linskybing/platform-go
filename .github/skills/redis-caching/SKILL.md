---
name: redis-caching
description: Redis caching patterns, distributed cache management, cache invalidation strategies, and performance optimization for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Redis Caching

This skill provides comprehensive guidelines for implementing Redis caching in platform-go for improved performance and scalability.

## When to Use

Apply this skill when:
- Caching frequently accessed data (user sessions, group storage lists, project configs)
- Implementing distributed caching across multiple application instances
- Reducing database queries and API calls
- Speeding up data retrieval for high-traffic endpoints
- Implementing real-time data synchronization
- Managing cache invalidation and TTL strategies
- Handling cache stampedes and concurrent access

## Quick Start: Using Cache Validation Scripts

This skill includes ready-to-use cache validation scripts:

```bash
# Validate Redis configuration and connectivity
bash .github/skills/redis-caching/scripts/validate-redis.sh

# Check cache hit ratios and performance metrics
bash .github/skills/redis-caching/scripts/cache-metrics.sh
```

## Cache Architecture

### 1. Redis Connection Pool

```go
import (
    "context"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// Initialize Redis client with connection pooling
func InitRedis(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         addr,                    // "localhost:6379"
        MaxRetries:   3,
        PoolSize:     10,
        MinIdleConns: 5,
        MaxConnAge:   time.Hour,
        
        // Connection timeout configuration
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        
        // Enable pipelining for bulk operations
        Protocol: 3,
    })
}

// Health check and connection validation
func HealthCheckRedis(ctx context.Context, client *redis.Client) error {
    status := client.Ping(ctx)
    if status.Err() != nil {
        return fmt.Errorf("redis health check failed: %w", status.Err())
    }
    return nil
}
```

### 2. Cache Key Design

```go
// Define consistent key patterns for different data types
const (
    // Format: cache:{entity}:{id}
    userSessionKey      = "cache:user:session:%s"
    groupStorageKey     = "cache:group:storage:%d"
    projectConfigKey    = "cache:project:config:%d"
    groupPVCListKey     = "cache:group:pvc:list:%d"
    
    // Tag-based keys for batch invalidation
    groupStorageTag     = "group:storage:%d"
    projectDataTag      = "project:data:%d"
)

// Generate cache key with group ID
func getGroupStorageCacheKey(groupID uint) string {
    return fmt.Sprintf("cache:group:storage:%d", groupID)
}

// Generate tag for cache invalidation
func getGroupStorageTag(groupID uint) string {
    return fmt.Sprintf("group:storage:%d", groupID)
}
```

### 3. Cache Operations with TTL

```go
type CacheService struct {
    client *redis.Client
    logger *slog.Logger
}

// Set stores data in cache with TTL
func (cs *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    start := time.Now()
    
    err := cs.client.Set(ctx, key, value, ttl).Err()
    if err != nil {
        cs.logger.Error("failed to set cache",
            "key", key,
            "ttl_seconds", ttl.Seconds(),
            "error", err)
        return fmt.Errorf("redis set failed: %w", err)
    }
    
    cs.logger.Debug("cache set",
        "key", key,
        "ttl_seconds", ttl.Seconds(),
        "duration_ms", time.Since(start).Milliseconds())
    
    return nil
}

// Get retrieves data from cache
func (cs *CacheService) Get(ctx context.Context, key string) (string, error) {
    start := time.Now()
    
    val, err := cs.client.Get(ctx, key).Result()
    if err == redis.Nil {
        cs.logger.Debug("cache miss", "key", key)
        return "", ErrCacheMiss
    }
    if err != nil {
        cs.logger.Error("failed to get cache",
            "key", key,
            "error", err)
        return "", fmt.Errorf("redis get failed: %w", err)
    }
    
    cs.logger.Debug("cache hit",
        "key", key,
        "duration_ms", time.Since(start).Milliseconds())
    
    return val, nil
}

// GetJSON retrieves and unmarshals JSON from cache
func (cs *CacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
    data, err := cs.Get(ctx, key)
    if err != nil {
        return err
    }
    
    if err := json.Unmarshal([]byte(data), dest); err != nil {
        cs.logger.Error("failed to unmarshal cached data",
            "key", key,
            "error", err)
        return fmt.Errorf("json unmarshal failed: %w", err)
    }
    
    return nil
}

// SetJSON marshals and stores JSON in cache
func (cs *CacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        cs.logger.Error("failed to marshal data for cache",
            "key", key,
            "error", err)
        return fmt.Errorf("json marshal failed: %w", err)
    }
    
    return cs.Set(ctx, key, string(data), ttl)
}
```

### 4. Cache Invalidation Strategies

```go
// Pattern-based invalidation (delete all matching keys)
func (cs *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
    start := time.Now()
    
    keys, err := cs.client.Keys(ctx, pattern).Result()
    if err != nil {
        cs.logger.Error("failed to find keys for invalidation",
            "pattern", pattern,
            "error", err)
        return fmt.Errorf("keys scan failed: %w", err)
    }
    
    if len(keys) == 0 {
        return nil
    }
    
    if err := cs.client.Del(ctx, keys...).Err(); err != nil {
        cs.logger.Error("failed to invalidate cache",
            "pattern", pattern,
            "count", len(keys),
            "error", err)
        return fmt.Errorf("cache invalidation failed: %w", err)
    }
    
    cs.logger.Info("cache invalidated",
        "pattern", pattern,
        "invalidated_keys", len(keys),
        "duration_ms", time.Since(start).Milliseconds())
    
    return nil
}

// Tag-based invalidation (atomic operation)
func (cs *CacheService) InvalidateByTag(ctx context.Context, tag string) error {
    return cs.InvalidatePattern(ctx, fmt.Sprintf("*:%s:*", tag))
}

// Single key invalidation
func (cs *CacheService) Invalidate(ctx context.Context, key string) error {
    err := cs.client.Del(ctx, key).Err()
    if err != nil && err != redis.Nil {
        cs.logger.Error("failed to invalidate key",
            "key", key,
            "error", err)
        return fmt.Errorf("cache delete failed: %w", err)
    }
    
    cs.logger.Debug("cache invalidated", "key", key)
    return nil
}
```

### 5. Cache-Aside Pattern (Most Common)

```go
// GetOrFetch implements cache-aside pattern with fallback to database
func (cs *CacheService) GetOrFetch(ctx context.Context, key string, ttl time.Duration, 
    fetchFunc func(context.Context) (interface{}, error)) (interface{}, error) {
    
    // Try cache first
    cached, err := cs.Get(ctx, key)
    if err == nil {
        cs.logger.Debug("cache hit", "key", key)
        return cached, nil
    }
    
    if err != ErrCacheMiss {
        cs.logger.Warn("cache read error, falling back to fetch",
            "key", key,
            "error", err)
    }
    
    // Cache miss or error, fetch from source
    cs.logger.Debug("cache miss, fetching from source", "key", key)
    data, err := fetchFunc(ctx)
    if err != nil {
        return nil, fmt.Errorf("fetch failed: %w", err)
    }
    
    // Store in cache (best effort, don't fail if cache write fails)
    if err := cs.SetJSON(ctx, key, data, ttl); err != nil {
        cs.logger.Warn("failed to cache data, but returning result",
            "key", key,
            "error", err)
    }
    
    return data, nil
}
```

### 6. Group Storage Cache Example

```go
// CachedStorageManager extends StorageManager with Redis caching
type CachedStorageManager struct {
    sm    *StorageManager
    cache *CacheService
}

// ListGroupPVCsCached retrieves PVCs with 2-level caching (memory + Redis)
func (csm *CachedStorageManager) ListGroupPVCsCached(ctx context.Context, groupID uint) ([]storage.GroupPVC, error) {
    cacheKey := fmt.Sprintf("cache:group:pvc:list:%d", groupID)
    
    // Try Redis cache first
    var pvcs []storage.GroupPVC
    err := csm.cache.GetJSON(ctx, cacheKey, &pvcs)
    if err == nil {
        slog.Debug("PVC list retrieved from Redis cache",
            "group_id", groupID,
            "count", len(pvcs))
        return pvcs, nil
    }
    
    if err != ErrCacheMiss {
        slog.Warn("redis cache read error, falling back to database",
            "group_id", groupID,
            "error", err)
    }
    
    // Cache miss, fetch from K8s/database
    pvcs, err = csm.sm.ListGroupPVCs(ctx, groupID)
    if err != nil {
        return nil, fmt.Errorf("failed to list PVCs: %w", err)
    }
    
    // Store in Redis (5 minute TTL for eventual consistency)
    if err := csm.cache.SetJSON(ctx, cacheKey, pvcs, 5*time.Minute); err != nil {
        slog.Warn("failed to cache PVC list",
            "group_id", groupID,
            "error", err)
    }
    
    return pvcs, nil
}

// InvalidateGroupStorage invalidates all caches for a group
func (csm *CachedStorageManager) InvalidateGroupStorage(groupID uint) error {
    pattern := fmt.Sprintf("cache:group:pvc:list:%d", groupID)
    
    // Invalidate Redis cache
    if err := csm.cache.InvalidatePattern(context.Background(), pattern); err != nil {
        slog.Error("failed to invalidate redis cache",
            "group_id", groupID,
            "error", err)
        // Don't fail operation, but log error
    }
    
    // Invalidate in-memory cache
    csm.sm.invalidateCache(groupID)
    
    slog.Info("group storage cache invalidated", "group_id", groupID)
    return nil
}
```

### 7. Handling Cache Stampede

```go
// SetWithMutex prevents cache stampede with distributed locking
func (cs *CacheService) SetWithMutex(ctx context.Context, key string, ttl time.Duration,
    fetchFunc func(context.Context) (interface{}, error)) error {
    
    lockKey := fmt.Sprintf("lock:%s", key)
    lockValue := uuid.New().String()
    
    // Try to acquire lock (non-blocking)
    locked, err := cs.client.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
    if err != nil {
        return fmt.Errorf("lock check failed: %w", err)
    }
    
    if !locked {
        // Another process is fetching, wait for result
        for attempts := 0; attempts < 10; attempts++ {
            val, err := cs.Get(ctx, key)
            if err == nil {
                return nil // Cache was populated by other process
            }
            time.Sleep(100 * time.Millisecond)
        }
        return fmt.Errorf("timeout waiting for cache population")
    }
    
    // We have the lock, fetch and cache
    data, err := fetchFunc(ctx)
    if err != nil {
        cs.client.Del(ctx, lockKey) // Release lock on error
        return fmt.Errorf("fetch failed: %w", err)
    }
    
    // Store in cache
    if err := cs.SetJSON(ctx, key, data, ttl); err != nil {
        cs.client.Del(ctx, lockKey)
        return fmt.Errorf("cache set failed: %w", err)
    }
    
    // Release lock
    cs.client.Del(ctx, lockKey)
    return nil
}
```

## Cache Metrics and Monitoring

### 8. Cache Statistics

```go
type CacheMetrics struct {
    Hits        int64
    Misses      int64
    Evictions   int64
    KeyCount    int64
    MemoryUsage int64
}

// GetMetrics collects cache statistics
func (cs *CacheService) GetMetrics(ctx context.Context) (*CacheMetrics, error) {
    info := cs.client.Info(ctx, "stats", "memory", "keyspace")
    
    // Parse Redis info response for metrics
    // Implementation depends on specific metrics needed
    
    return &CacheMetrics{
        // Parse from info response
    }, nil
}

// LogMetrics logs cache performance metrics
func (cs *CacheService) LogMetrics(ctx context.Context) {
    metrics, err := cs.GetMetrics(ctx)
    if err != nil {
        cs.logger.Error("failed to get cache metrics", "error", err)
        return
    }
    
    hitRate := float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses) * 100
    
    slog.Info("cache metrics",
        "hits", metrics.Hits,
        "misses", metrics.Misses,
        "hit_rate_percent", hitRate,
        "evictions", metrics.Evictions,
        "keys", metrics.KeyCount,
        "memory_bytes", metrics.MemoryUsage)
}
```

## Best Practices

### Do's ✓
- Use appropriate TTL based on data freshness requirements
- Implement cache invalidation on data updates
- Monitor cache hit ratios and adjust strategies
- Use key prefixes for logical grouping
- Implement proper error handling and fallbacks
- Log cache operations for debugging
- Use connection pooling
- Validate cache data before use

### Don'ts ✗
- Never cache sensitive data (passwords, tokens, PII)
- Don't rely solely on cache without fallback to database
- Avoid very short TTLs that increase cache misses
- Don't cache unbounded amounts of data
- Avoid complex nested structures in cache
- Don't ignore cache failures (fail gracefully)

## Configuration Examples

```yaml
# Docker Compose Redis service
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

# Kubernetes Redis Helm values
redis:
  enabled: true
  auth:
    enabled: true
    password: "secure-password"
  master:
    persistence:
      enabled: true
      size: 10Gi
  maxmemory: "512Mi"
  maxmemoryPolicy: "allkeys-lru"
```

## See Also

- [Database Best Practices](../database-best-practices/SKILL.md)
- [Monitoring & Observability](../monitoring-observability/SKILL.md)
- [Production Readiness Checklist](../production-readiness-checklist/SKILL.md)
