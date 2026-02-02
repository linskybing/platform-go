package k8s

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/repository"
)

const (
	// PVC ID generation constants
	uuidShortLength = 8
	
	// Namespace naming patterns
	groupNamespacePattern = "group-%d-storage"
	groupPVCIDPattern     = "group-%d-%s"
)

type StorageManager struct {
	repos      *repository.Repos
	pvcCache   map[uint]*CacheEntry // Updated type to use CacheEntry
	cacheMutex sync.RWMutex
}

func NewStorageManager(repos *repository.Repos) *StorageManager {
	return &StorageManager{
		repos:    repos,
		pvcCache: make(map[uint]*CacheEntry),
	}
}

func (sm *StorageManager) generateGroupPVCID(groupID uint) string {
	return fmt.Sprintf(groupPVCIDPattern, groupID, uuid.New().String()[:uuidShortLength])
}

func (sm *StorageManager) getGroupNamespace(groupID uint) string {
	return fmt.Sprintf(groupNamespacePattern, groupID)
}