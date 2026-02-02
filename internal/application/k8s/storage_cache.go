package k8s

import (
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
func (sm *StorageManager) getCachedPVCs(groupID uint) ([]storage.GroupPVC, bool) {
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
func (sm *StorageManager) setCachedPVCs(groupID uint, pvcs []storage.GroupPVC, ttl time.Duration) {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	sm.pvcCache[groupID] = &CacheEntry{
		PVCs:      pvcs,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	slog.Debug("cached PVCs for group", "group_id", groupID, "count", len(pvcs), "ttl_seconds", ttl.Seconds())
}

// invalidateCache removes cached PVCs for a specific group.
func (sm *StorageManager) invalidateCache(groupID uint) {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	delete(sm.pvcCache, groupID)
	slog.Debug("invalidated cache for group", "group_id", groupID)
}

// invalidateAllCache clears all cached PVCs.
func (sm *StorageManager) invalidateAllCache() {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	size := len(sm.pvcCache)
	sm.pvcCache = make(map[uint]*CacheEntry)
	slog.Debug("cleared all PVC cache", "cleared_entries", size)
}

// GetCacheStats returns cache statistics for monitoring.
func (sm *StorageManager) GetCacheStats() map[string]interface{} {
	sm.cacheMutex.RLock()
	defer sm.cacheMutex.RUnlock()

	totalCached := 0
	for _, entry := range sm.pvcCache {
		totalCached += len(entry.PVCs)
	}

	return map[string]interface{}{
		"cached_groups": len(sm.pvcCache),
		"total_pvcs":    totalCached,
	}
}
