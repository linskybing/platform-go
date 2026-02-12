package k8s

import (
	"fmt"

	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

// ImageValidator defines the contract used to validate image access by project.
type ImageValidator interface {
	ValidateImageForProject(name, tag string, projectID *string) (bool, error)
}

// K8sService orchestrates Kubernetes-related operations.
// It delegates to specialized managers for specific concerns.
type K8sService struct {
	repos              *repository.Repos
	StorageManager     *StorageManager
	FileBrowserManager *FileBrowserManager
	PVCBindingManager  *PVCBindingManager
	PermissionManager  *PermissionManager
	userStorageManager *UserStorageManager
}

// NewK8sService creates a new K8sService with all manager dependencies initialized.
func NewK8sService(repos *repository.Repos, imageValidator ImageValidator, cacheSvc *cache.Service) (*K8sService, error) {
	if repos == nil {
		return nil, fmt.Errorf("failed to initialize K8sService: %w", ErrNilRepository)
	}
	if imageValidator == nil {
		return nil, fmt.Errorf("failed to initialize K8sService: %w", ErrNilRequest)
	}

	// Initialize permission manager first so it can be passed into FileBrowserManager
	permMgr := NewPermissionManager(repos)

	service := &K8sService{
		repos:              repos,
		PermissionManager:  permMgr,
		FileBrowserManager: NewFileBrowserManager(permMgr),
		PVCBindingManager:  NewPVCBindingManager(repos, cacheSvc),
		userStorageManager: NewUserStorageManager(repos.Storage, repos.User),
		StorageManager:     NewStorageManager(repos, cacheSvc),
	}

	return service, nil
}
