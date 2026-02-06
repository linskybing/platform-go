package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/domain/storage"
)

// CacheEntry holds PVC data and metadata
type CacheEntry struct {
	PVCs      []storage.GroupPVC
	Timestamp time.Time
	TTL       time.Duration
}

// IsExpired checks if cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Since(ce.Timestamp) > ce.TTL
}

// getCachedPVCs retrieves cached PVCs for a group with TTL validation.
// Returns the cached PVCs and a boolean indicating cache hit.
func (sm *StorageManager) getCachedPVCs(groupID string) ([]storage.GroupPVC, bool) {
	sm.cacheMutex.RLock()
	defer sm.cacheMutex.RUnlock()

	entry, ok := sm.pvcCache[groupID]
	if !ok {
		slog.Debug("cache miss for group", "group_id", groupID)
		return nil, false
	}

	if entry.IsExpired() {
		slog.Debug("cache expired for group", "group_id", groupID, "expired_at", entry.Timestamp.Add(entry.TTL))
		return nil, false
	}

	slog.Debug("cache hit for group", "group_id", groupID, "cached_pvcs", len(entry.PVCs))
	return entry.PVCs, true
}

// setCachedPVCs stores PVCs in cache with TTL.
func (sm *StorageManager) setCachedPVCs(groupID string, pvcs []storage.GroupPVC, ttl time.Duration) {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	sm.pvcCache[groupID] = &CacheEntry{
		PVCs:      pvcs,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
	slog.Debug("cache set for group", "group_id", groupID, "count", len(pvcs))
}

// invalidatePVCsCache removes cached PVCs for a group
func (sm *StorageManager) invalidatePVCsCache(ctx context.Context, groupID string) {
	sm.cacheMutex.Lock()
	delete(sm.pvcCache, groupID)
	sm.cacheMutex.Unlock()

	// Note: Cache service uses TTL-based expiration, manual deletion not supported
	slog.Debug("cache invalidated for group", "group_id", groupID)
}

func (sm *StorageManager) pvcCacheKey(groupID string) string {
	return fmt.Sprintf("group:%s:pvcs", groupID)
}
