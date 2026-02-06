package repository

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"gorm.io/gorm"
)

// ProjectPVCBindingRepo defines project PVC binding data access interface
type ProjectPVCBindingRepo interface {
	// ProjectPVCBinding operations
	CreateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error
	GetBinding(ctx context.Context, id string) (*storage.ProjectPVCBinding, error)
	GetBindingByProjectPVC(ctx context.Context, namespace, pvcName string) (*storage.ProjectPVCBinding, error)
	ListBindings(ctx context.Context, projectID string) ([]storage.ProjectPVCBinding, error)
	ListBindingsByUser(ctx context.Context, userID string) ([]storage.ProjectPVCBinding, error)
	ListBindingsByGroupPVC(ctx context.Context, groupPVCID string) ([]storage.ProjectPVCBinding, error)
	UpdateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error
	UpdateBindingStatus(ctx context.Context, id string, status string) error
	DeleteBinding(ctx context.Context, id string) error
	DeleteBindingByProjectPVC(ctx context.Context, namespace, pvcName string) error

	WithTx(tx *gorm.DB) ProjectPVCBindingRepo
}

// ProjectPVCBindingRepoImpl implements ProjectPVCBindingRepo
type ProjectPVCBindingRepoImpl struct {
	db *gorm.DB
}

// NewProjectPVCBindingRepo creates a new ProjectPVCBindingRepo
func NewProjectPVCBindingRepo(db *gorm.DB) ProjectPVCBindingRepo {
	return &ProjectPVCBindingRepoImpl{db: db}
}

// CreateBinding creates a new binding record
func (r *ProjectPVCBindingRepoImpl) CreateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error {
	if err := r.db.WithContext(ctx).Create(binding).Error; err != nil {
		return fmt.Errorf("failed to create binding: %w", err)
	}
	return nil
}

// GetBinding retrieves a binding by ID
func (r *ProjectPVCBindingRepoImpl) GetBinding(ctx context.Context, id string) (*storage.ProjectPVCBinding, error) {
	var binding storage.ProjectPVCBinding
	err := r.db.WithContext(ctx).First(&binding, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetBindingByProjectPVC retrieves a binding by project namespace and PVC name
func (r *ProjectPVCBindingRepoImpl) GetBindingByProjectPVC(ctx context.Context, namespace, pvcName string) (*storage.ProjectPVCBinding, error) {
	var binding storage.ProjectPVCBinding
	err := r.db.WithContext(ctx).
		Where("project_namespace = ? AND project_pvc_name = ?", namespace, pvcName).
		First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// ListBindings retrieves all bindings for a project
func (r *ProjectPVCBindingRepoImpl) ListBindings(ctx context.Context, projectID string) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&bindings).Error
	return bindings, err
}

// ListBindingsByUser retrieves all bindings for a user
func (r *ProjectPVCBindingRepoImpl) ListBindingsByUser(ctx context.Context, userID string) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&bindings).Error
	return bindings, err
}

// ListBindingsByGroupPVC lists all bindings for a group PVC
func (r *ProjectPVCBindingRepoImpl) ListBindingsByGroupPVC(ctx context.Context, groupPVCID string) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	err := r.db.WithContext(ctx).
		Where("group_pvc_id = ?", groupPVCID).
		Find(&bindings).Error
	return bindings, err
}

// UpdateBinding updates a binding
func (r *ProjectPVCBindingRepoImpl) UpdateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error {
	if err := r.db.WithContext(ctx).Save(binding).Error; err != nil {
		return fmt.Errorf("failed to update binding: %w", err)
	}
	return nil
}

// UpdateBindingStatus updates binding status
func (r *ProjectPVCBindingRepoImpl) UpdateBindingStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).
		Model(&storage.ProjectPVCBinding{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// DeleteBinding deletes a binding record
func (r *ProjectPVCBindingRepoImpl) DeleteBinding(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Delete(&storage.ProjectPVCBinding{}, "id = ?", id).Error
}

// DeleteBindingByProjectPVC deletes a binding by project PVC
func (r *ProjectPVCBindingRepoImpl) DeleteBindingByProjectPVC(ctx context.Context, namespace, pvcName string) error {
	return r.db.WithContext(ctx).
		Delete(&storage.ProjectPVCBinding{}, "project_namespace = ? AND project_pvc_name = ?", namespace, pvcName).Error
}

func (r *ProjectPVCBindingRepoImpl) WithTx(tx *gorm.DB) ProjectPVCBindingRepo {
	if tx == nil {
		return r
	}
	return &ProjectPVCBindingRepoImpl{db: tx}
}
