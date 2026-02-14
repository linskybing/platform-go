package k8s

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
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
	return fmt.Sprintf("%s-%s", prefix, randomString(10)), nil
}

func (sm *StorageManager) generateGroupPVCID(groupID string) (string, error) {
	return fmt.Sprintf("group-%s-%s", groupID, randomString(10)), nil
}

func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
