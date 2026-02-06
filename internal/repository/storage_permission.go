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
	GetPermission(ctx context.Context, groupID, userID string, pvcID string) (*storage.GroupStoragePermission, error)
	ListPermissions(ctx context.Context, groupID string) ([]storage.GroupStoragePermission, error)
	ListUserPermissions(ctx context.Context, userID string) ([]storage.GroupStoragePermission, error)
	UpdatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error
	RevokePermission(ctx context.Context, id string) error
	DeletePermission(ctx context.Context, id string) error

	// GroupStorageAccessPolicy operations
	CreateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error
	GetAccessPolicy(ctx context.Context, pvcID string) (*storage.GroupStorageAccessPolicy, error)
	ListAccessPolicies(ctx context.Context, groupID string) ([]storage.GroupStorageAccessPolicy, error)
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
func (r *StoragePermissionRepoImpl) GetPermission(ctx context.Context, groupID, userID string, pvcID string) (*storage.GroupStoragePermission, error) {
	var perm storage.GroupStoragePermission
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND pvc_id = ?", groupID, userID, pvcID).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// ListPermissions retrieves all permissions for a group
func (r *StoragePermissionRepoImpl) ListPermissions(ctx context.Context, groupID string) ([]storage.GroupStoragePermission, error) {
	var perms []storage.GroupStoragePermission
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&perms).Error
	return perms, err
}

// ListUserPermissions retrieves all permissions for a user
func (r *StoragePermissionRepoImpl) ListUserPermissions(ctx context.Context, userID string) ([]storage.GroupStoragePermission, error) {
	var perms []storage.GroupStoragePermission
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&perms).Error
	return perms, err
}

// UpdatePermission updates an existing permission
func (r *StoragePermissionRepoImpl) UpdatePermission(ctx context.Context, perm *storage.GroupStoragePermission) error {
	if err := r.db.WithContext(ctx).Save(perm).Error; err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}
	return nil
}

// RevokePermission revokes a permission
func (r *StoragePermissionRepoImpl) RevokePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&storage.GroupStoragePermission{}).
		Where("id = ?", id).
		Update("revoked_at", gorm.Expr("NOW()")).Error
}

// DeletePermission deletes a permission record
func (r *StoragePermissionRepoImpl) DeletePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Delete(&storage.GroupStoragePermission{}, "id = ?", id).Error
}

// CreateAccessPolicy creates a new access policy
func (r *StoragePermissionRepoImpl) CreateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error {
	if err := r.db.WithContext(ctx).Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create access policy: %w", err)
	}
	return nil
}

// GetAccessPolicy retrieves the access policy for a PVC
func (r *StoragePermissionRepoImpl) GetAccessPolicy(ctx context.Context, pvcID string) (*storage.GroupStorageAccessPolicy, error) {
	var policy storage.GroupStorageAccessPolicy
	err := r.db.WithContext(ctx).
		Where("pvc_id = ?", pvcID).
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListAccessPolicies retrieves all access policies for a group
func (r *StoragePermissionRepoImpl) ListAccessPolicies(ctx context.Context, groupID string) ([]storage.GroupStorageAccessPolicy, error) {
	var policies []storage.GroupStorageAccessPolicy
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&policies).Error
	return policies, err
}

// UpdateAccessPolicy updates an access policy
func (r *StoragePermissionRepoImpl) UpdateAccessPolicy(ctx context.Context, policy *storage.GroupStorageAccessPolicy) error {
	if err := r.db.WithContext(ctx).Save(policy).Error; err != nil {
		return fmt.Errorf("failed to update access policy: %w", err)
	}
	return nil
}

// DeleteAccessPolicy deletes an access policy
func (r *StoragePermissionRepoImpl) DeleteAccessPolicy(ctx context.Context, pvcID string) error {
	return r.db.WithContext(ctx).
		Delete(&storage.GroupStorageAccessPolicy{}, "pvc_id = ?", pvcID).Error
}

func (r *StoragePermissionRepoImpl) WithTx(tx *gorm.DB) StoragePermissionRepo {
	if tx == nil {
		return r
	}
	return &StoragePermissionRepoImpl{db: tx}
}
