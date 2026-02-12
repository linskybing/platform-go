package k8s

import (
	"fmt"
	"sync"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// Namespace naming patterns
	groupNamespacePattern = "group-%s-storage"
	userNamespacePattern  = "user-%s-storage"
)

// StorageManager orchestrates storage-related operations between Kubernetes and the database.
type StorageManager struct {
	repos       *repository.Repos
	storageRepo storage.StorageRepo
	cache       *cache.Service
	pvcCache    map[string]*CacheEntry
	cacheMutex  sync.RWMutex
}

// NewStorageManager creates a new StorageManager.
func NewStorageManager(repos *repository.Repos, cacheSvc *cache.Service) *StorageManager {
	var storageRepo storage.StorageRepo
	if repos != nil {
		storageRepo = repos.Storage
	}
	return &StorageManager{
		repos:       repos,
		storageRepo: storageRepo,
		cache:       cacheSvc,
		pvcCache:    make(map[string]*CacheEntry),
	}
}

// --- Namespace Helpers ---

func (sm *StorageManager) getGroupNamespace(groupID string) string {
	return fmt.Sprintf(groupNamespacePattern, groupID)
}

func (sm *StorageManager) getUserNamespace(userID string) string {
	safeUser := k8sclient.ToSafeK8sName(userID)
	return fmt.Sprintf(config.UserStorageNs, safeUser)
}

// --- ID Generation ---

func (sm *StorageManager) generatePVCName(prefix string) (string, error) {
	suffix, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 10)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", prefix, suffix), nil
}

func (sm *StorageManager) generateGroupPVCID(groupID string) (string, error) {
	suffix, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 10)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("group-%s-%s", groupID, suffix), nil
}
