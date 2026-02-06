package k8s

import (
	"fmt"
	"sync"

	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// Namespace naming patterns
	groupNamespacePattern = "group-%s-storage"
)

type StorageManager struct {
	repos      *repository.Repos
	cache      *cache.Service
	pvcCache   map[string]*CacheEntry // Updated to use string keys
	cacheMutex sync.RWMutex
}

func NewStorageManager(repos *repository.Repos, cacheSvc *cache.Service) *StorageManager {
	return &StorageManager{
		repos:    repos,
		cache:    cacheSvc,
		pvcCache: make(map[string]*CacheEntry),
	}
}

func (sm *StorageManager) generateGroupPVCID(groupID string) (string, error) {
	// Generate a short NanoID for the suffix
	suffix, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 10)
	if err != nil {
		return "", err
	}
	// Return format: group-{gid}-{suffix}
	return fmt.Sprintf("group-%s-%s", groupID, suffix), nil
}

func (sm *StorageManager) getGroupNamespace(groupID string) string {
	return fmt.Sprintf(groupNamespacePattern, groupID)
}
