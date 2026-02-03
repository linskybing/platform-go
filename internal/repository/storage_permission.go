package repository

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"gorm.io/gorm"
)

// StoragePermissionRepo defines storage permission data access interface
type StoragePermissionRepo interface {
	// GroupStoragePermission operations
	CreatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error
	GetPermission(ctx context.Context, groupID, userID uint, pvcID string) (*storage.GroupStoragePermission, error)
	ListPermissions(ctx context.Context, groupID uint) ([]storage.GroupStoragePermission, error)
	ListUserPermissions(ctx context.Context, userID uint) ([]storage.GroupStoragePermission, error)
	UpdatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error
	RevokePermission(ctx context.Context, id uint) error
	DeletePermission(ctx context.Context, id uint) error

	// GroupStorageAccessPolicy operations
	CreateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error
	GetAccessPolicy(ctx context.Context, pvcID string) (*storage.GroupStorageAccessPolicy, error)
	ListAccessPolicies(ctx context.Context, groupID uint) ([]storage.GroupStorageAccessPolicy, error)
	UpdateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error
	DeleteAccessPolicy(ctx context.Context, pvcID string) error

	WithTx(tx *gorm.DB) StoragePermissionRepo
}

// StoragePermissionRepoImpl implements StoragePermissionRepo
type StoragePermissionRepoImpl struct {
	db *gorm.DB
}

// NewStoragePermissionRepo creates a new StoragePermissionRepo
func NewStoragePermissionRepo(db *gorm.DB) StoragePermissionRepo {
	return &StoragePermissionRepoImpl{db: db}
}

// CreatePermission creates a new group storage permission
func (r *StoragePermissionRepoImpl) CreatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error {
	if err := r.db.WithContext(ctx).Create(perm).Error; err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

// GetPermission retrieves a specific permission by group, user, and PVC
func (r *StoragePermissionRepoImpl) GetPermission(ctx context.Context, groupID, userID uint, pvcID string) (*storage.GroupStoragePermission, error) {
	var perm storage.GroupStoragePermission
	if err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND pvc_id = ? AND revoked_at IS NULL", groupID, userID, pvcID).
		First(&perm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &perm, nil
}

// ListPermissions lists all active permissions for a group
func (r *StoragePermissionRepoImpl) ListPermissions(ctx context.Context, groupID uint) ([]storage.GroupStoragePermission, error) {
	var perms []storage.GroupStoragePermission
	if err := r.db.WithContext(ctx).
		Where("group_id = ? AND revoked_at IS NULL", groupID).
		Order("created_at DESC").
		Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return perms, nil
}

// ListUserPermissions lists all active permissions for a user
func (r *StoragePermissionRepoImpl) ListUserPermissions(ctx context.Context, userID uint) ([]storage.GroupStoragePermission, error) {
	var perms []storage.GroupStoragePermission
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at DESC").
		Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("failed to list user permissions: %w", err)
	}
	return perms, nil
}

// UpdatePermission updates an existing permission
func (r *StoragePermissionRepoImpl) UpdatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error {
	if err := r.db.WithContext(ctx).Model(perm).Updates(perm).Error; err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}
	return nil
}

// RevokePermission marks a permission as revoked
func (r *StoragePermissionRepoImpl) RevokePermission(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).
		Model(&storage.GroupStoragePermission{}).
		Where("id = ?", id).
		Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}
	return nil
}

// DeletePermission hard-deletes a permission record
func (r *StoragePermissionRepoImpl) DeletePermission(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&storage.GroupStoragePermission{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// CreateAccessPolicy creates a new access policy
func (r *StoragePermissionRepoImpl) CreateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error {
	if err := r.db.WithContext(ctx).Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create access policy: %w", err)
	}
	return nil
}

// GetAccessPolicy retrieves an access policy by PVC ID
func (r *StoragePermissionRepoImpl) GetAccessPolicy(ctx context.Context, pvcID string) (*storage.GroupStorageAccessPolicy, error) {
	var policy storage.GroupStorageAccessPolicy
	if err := r.db.WithContext(ctx).
		Where("pvc_id = ?", pvcID).
		First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("access policy not found")
		}
		return nil, fmt.Errorf("failed to get access policy: %w", err)
	}
	return &policy, nil
}

// ListAccessPolicies lists all access policies for a group
func (r *StoragePermissionRepoImpl) ListAccessPolicies(ctx context.Context, groupID uint) ([]storage.GroupStorageAccessPolicy, error) {
	var policies []storage.GroupStorageAccessPolicy
	if err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list access policies: %w", err)
	}
	return policies, nil
}

// UpdateAccessPolicy updates an existing access policy
func (r *StoragePermissionRepoImpl) UpdateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error {
	if err := r.db.WithContext(ctx).Model(policy).Updates(policy).Error; err != nil {
		return fmt.Errorf("failed to update access policy: %w", err)
	}
	return nil
}

// DeleteAccessPolicy deletes an access policy
func (r *StoragePermissionRepoImpl) DeleteAccessPolicy(ctx context.Context, pvcID string) error {
	if err := r.db.WithContext(ctx).Where("pvc_id = ?", pvcID).Delete(&storage.GroupStorageAccessPolicy{}).Error; err != nil {
		return fmt.Errorf("failed to delete access policy: %w", err)
	}
	return nil
}

// WithTx returns a repository using a transaction
func (r *StoragePermissionRepoImpl) WithTx(tx *gorm.DB) StoragePermissionRepo {
	return &StoragePermissionRepoImpl{db: tx}
}
