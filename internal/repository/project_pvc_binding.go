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
	GetBinding(ctx context.Context, id uint) (*storage.ProjectPVCBinding, error)
	GetBindingByProjectPVC(ctx context.Context, namespace, pvcName string) (*storage.ProjectPVCBinding, error)
	ListBindings(ctx context.Context, projectID uint) ([]storage.ProjectPVCBinding, error)
	ListBindingsByUser(ctx context.Context, userID uint) ([]storage.ProjectPVCBinding, error)
	ListBindingsByGroupPVC(ctx context.Context, groupPVCID string) ([]storage.ProjectPVCBinding, error)
	UpdateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error
	UpdateBindingStatus(ctx context.Context, id uint, status string) error
	DeleteBinding(ctx context.Context, id uint) error
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

// CreateBinding creates a new project PVC binding
func (r *ProjectPVCBindingRepoImpl) CreateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error {
	if err := r.db.WithContext(ctx).Create(binding).Error; err != nil {
		return fmt.Errorf("failed to create binding: %w", err)
	}
	return nil
}

// GetBinding retrieves a binding by ID
func (r *ProjectPVCBindingRepoImpl) GetBinding(ctx context.Context, id uint) (*storage.ProjectPVCBinding, error) {
	var binding storage.ProjectPVCBinding
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("binding not found")
		}
		return nil, fmt.Errorf("failed to get binding: %w", err)
	}
	return &binding, nil
}

// GetBindingByProjectPVC retrieves a binding by project namespace and PVC name
func (r *ProjectPVCBindingRepoImpl) GetBindingByProjectPVC(ctx context.Context, namespace, pvcName string) (*storage.ProjectPVCBinding, error) {
	var binding storage.ProjectPVCBinding
	if err := r.db.WithContext(ctx).
		Where("project_namespace = ? AND project_pvc_name = ?", namespace, pvcName).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("binding not found")
		}
		return nil, fmt.Errorf("failed to get binding: %w", err)
	}
	return &binding, nil
}

// ListBindings lists all bindings for a project
func (r *ProjectPVCBindingRepoImpl) ListBindings(ctx context.Context, projectID uint) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}
	return bindings, nil
}

// ListBindingsByUser lists all bindings created by a user
func (r *ProjectPVCBindingRepoImpl) ListBindingsByUser(ctx context.Context, userID uint) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to list user bindings: %w", err)
	}
	return bindings, nil
}

// ListBindingsByGroupPVC lists all bindings for a group PVC
func (r *ProjectPVCBindingRepoImpl) ListBindingsByGroupPVC(ctx context.Context, groupPVCID string) ([]storage.ProjectPVCBinding, error) {
	var bindings []storage.ProjectPVCBinding
	if err := r.db.WithContext(ctx).
		Where("group_pvc_id = ?", groupPVCID).
		Order("created_at DESC").
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to list group PVC bindings: %w", err)
	}
	return bindings, nil
}

// UpdateBinding updates an existing binding
func (r *ProjectPVCBindingRepoImpl) UpdateBinding(ctx context.Context, binding *storage.ProjectPVCBinding) error {
	if err := r.db.WithContext(ctx).Model(binding).Updates(binding).Error; err != nil {
		return fmt.Errorf("failed to update binding: %w", err)
	}
	return nil
}

// UpdateBindingStatus updates only the status of a binding
func (r *ProjectPVCBindingRepoImpl) UpdateBindingStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).
		Model(&storage.ProjectPVCBinding{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update binding status: %w", err)
	}
	return nil
}

// DeleteBinding deletes a binding by ID
func (r *ProjectPVCBindingRepoImpl) DeleteBinding(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&storage.ProjectPVCBinding{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete binding: %w", err)
	}
	return nil
}

// DeleteBindingByProjectPVC deletes a binding by project namespace and PVC name
func (r *ProjectPVCBindingRepoImpl) DeleteBindingByProjectPVC(ctx context.Context, namespace, pvcName string) error {
	if err := r.db.WithContext(ctx).
		Where("project_namespace = ? AND project_pvc_name = ?", namespace, pvcName).
		Delete(&storage.ProjectPVCBinding{}).Error; err != nil {
		return fmt.Errorf("failed to delete binding: %w", err)
	}
	return nil
}

// WithTx returns a repository using a transaction
func (r *ProjectPVCBindingRepoImpl) WithTx(tx *gorm.DB) ProjectPVCBindingRepo {
	return &ProjectPVCBindingRepoImpl{db: tx}
}
