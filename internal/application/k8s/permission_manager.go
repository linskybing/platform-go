package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/logger"
)

// PermissionManager handles storage permission operations
type PermissionManager struct {
	repos *repository.Repos
}

// NewPermissionManager creates a new PermissionManager
func NewPermissionManager(repos *repository.Repos) *PermissionManager {
	return &PermissionManager{
		repos: repos,
	}
}

// SetPermission sets or updates a user's permission for a group PVC
func (pm *PermissionManager) SetPermission(ctx context.Context, req *storage.SetStoragePermissionRequest, granterID string) error {
	// Try to get existing permission
	existingPerm, err := pm.repos.StoragePermission.GetPermission(ctx, req.GroupID, req.UserID, req.PVCID)

	if err == nil {
		// Permission exists, update it
		existingPerm.Permission = req.Permission
		existingPerm.GrantedBy = granterID
		existingPerm.RevokedAt = nil // Un-revoke if it was revoked

		if err := pm.repos.StoragePermission.UpdatePermission(ctx, existingPerm); err != nil {
			return fmt.Errorf("failed to update permission: %w", err)
		}

		logger.Info("updated storage permission",
			"group_id", req.GroupID,
			"pvc_id", req.PVCID,
			"user_id", req.UserID,
			"permission", req.Permission)
	} else {
		// Permission doesn't exist, create it
		perm := &storage.GroupStoragePermission{
			GroupID:    req.GroupID,
			PVCID:      req.PVCID,
			UserID:     req.UserID,
			Permission: req.Permission,
			GrantedBy:  granterID,
			GrantedAt:  time.Now(),
		}

		if err := pm.repos.StoragePermission.CreatePermission(ctx, perm); err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}

		logger.Info("created storage permission",
			"group_id", req.GroupID,
			"pvc_id", req.PVCID,
			"user_id", req.UserID,
			"permission", req.Permission)
	}

	return nil
}

// BatchSetPermissions sets permissions for multiple users
func (pm *PermissionManager) BatchSetPermissions(ctx context.Context, req *storage.BatchSetPermissionsRequest, granterID string) error {
	for _, userPerm := range req.Permissions {
		permReq := &storage.SetStoragePermissionRequest{
			GroupID:    req.GroupID,
			PVCID:      req.PVCID,
			UserID:     userPerm.UserID,
			Permission: userPerm.Permission,
		}
		if err := pm.SetPermission(ctx, permReq, granterID); err != nil {
			logger.Error("failed to set permission", "user_id", userPerm.UserID, "error", err)
			continue
		}
	}
	return nil
}

// GetUserPermission retrieves a user's permission for a specific PVC
func (pm *PermissionManager) GetUserPermission(ctx context.Context, userID, groupID string, pvcID string) (*storage.GroupStoragePermission, error) {
	perm, err := pm.repos.StoragePermission.GetPermission(ctx, groupID, userID, pvcID)
	if err != nil {
		return &storage.GroupStoragePermission{
			UserID:     userID,
			GroupID:    groupID,
			PVCID:      pvcID,
			Permission: storage.PermissionNone,
		}, nil
	}
	return perm, nil
}

// SetAccessPolicy sets the default access policy for a group PVC
func (pm *PermissionManager) SetAccessPolicy(ctx context.Context, req *storage.SetStorageAccessPolicyRequest, adminID string) error {
	policy := &storage.GroupStorageAccessPolicy{
		GroupID:           req.GroupID,
		PVCID:             req.PVCID,
		DefaultPermission: req.DefaultPermission,
		AdminOnly:         req.AdminOnly,
		CreatedBy:         adminID,
	}

	if err := pm.repos.StoragePermission.CreateAccessPolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to set access policy: %w", err)
	}

	logger.Info("set storage access policy",
		"group_id", req.GroupID,
		"pvc_id", req.PVCID,
		"default_permission", req.DefaultPermission)

	return nil
}

// SyncPermissionsOnGroupChange syncs permissions when user joins/leaves group
func (pm *PermissionManager) SyncPermissionsOnGroupChange(ctx context.Context, userID, groupID uint, action string) error {
	logger.Info("syncing permissions on group change",
		"user_id", userID,
		"group_id", groupID,
		"action", action)
	return nil
}

// RevokePermission revokes a user's permission
func (pm *PermissionManager) RevokePermission(ctx context.Context, userID, groupID, pvcID string) error {
	perm, err := pm.repos.StoragePermission.GetPermission(ctx, groupID, userID, pvcID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if err := pm.repos.StoragePermission.RevokePermission(ctx, perm.ID); err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	logger.Info("revoked storage permission",
		"user_id", userID,
		"group_id", groupID,
		"pvc_id", pvcID)

	return nil
}

// ListGroupPermissions lists all permissions for a group
func (pm *PermissionManager) ListGroupPermissions(ctx context.Context, groupID string) ([]storage.GroupStoragePermission, error) {
	return pm.repos.StoragePermission.ListPermissions(ctx, groupID)
}

// ListPVCPermissions lists all active permissions for a specific PVC
func (pm *PermissionManager) ListPVCPermissions(ctx context.Context, groupID, pvcID string) ([]storage.GroupStoragePermission, error) {
	return pm.repos.StoragePermission.ListPermissionsByPVC(ctx, groupID, pvcID)
}

// ListUserPermissions lists all permissions for a user
func (pm *PermissionManager) ListUserPermissions(ctx context.Context, userID string) ([]storage.GroupStoragePermission, error) {
	return pm.repos.StoragePermission.ListUserPermissions(ctx, userID)
}
