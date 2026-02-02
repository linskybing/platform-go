package k8s

import (
	"fmt"

	"github.com/linskybing/platform-go/internal/repository"
)

// ImageValidator defines the contract used to validate image access by project.
type ImageValidator interface {
	ValidateImageForProject(name, tag string, projectID *uint) (bool, error)
}

// K8sService orchestrates Kubernetes-related operations.
// It delegates to specialized managers for specific concerns.
type K8sService struct {
	repos              *repository.Repos
	storageManager     *StorageManager
	fileBrowserManager *FileBrowserManager
	userStorageManager *UserStorageManager
}

// NewK8sService creates a new K8sService with all manager dependencies initialized.
func NewK8sService(repos *repository.Repos, imageValidator ImageValidator) (*K8sService, error) {
	if repos == nil {
		return nil, fmt.Errorf("failed to initialize K8sService: %w", ErrNilRepository)
	}
	if imageValidator == nil {
		return nil, fmt.Errorf("failed to initialize K8sService: %w", ErrNilRequest)
	}

	service := &K8sService{
		repos:              repos,
		fileBrowserManager: NewFileBrowserManager(),
		userStorageManager: NewUserStorageManager(),
		storageManager:     NewStorageManager(repos),
	}

	return service, nil
}

// Accessors for manager dependencies.
func (s *K8sService) StorageMgr() *StorageManager         { return s.storageManager }
func (s *K8sService) FileBrowserMgr() *FileBrowserManager { return s.fileBrowserManager }
func (s *K8sService) UserStorageMgr() *UserStorageManager { return s.userStorageManager }
