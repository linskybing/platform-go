package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/types"
	"gorm.io/gorm"
)

func IsSuperAdmin(uid string, repos repository.UserGroupRepo) (bool, error) {
	return repos.IsSuperAdmin(context.Background(), uid)
}

var GetUserIDFromContext = func(c *gin.Context) (string, error) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return "", fmt.Errorf("user claims not found in context: missing claims")
	}

	claims, ok := claimsVal.(*types.Claims)
	if !ok {
		return "", fmt.Errorf("invalid user claims type (expected *types.Claims): type assertion failed")
	}

	return claims.UserID, nil
}

var GetUserNameFromContext = func(c *gin.Context) (string, error) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return "", fmt.Errorf("user claims not found in context: missing claims")
	}

	claims, ok := claimsVal.(*types.Claims)
	if !ok {
		return "", fmt.Errorf("invalid user claims type (expected *types.Claims): type assertion failed")
	}

	return claims.Username, nil
}

func HasGroupRole(userID string, gid string, roles []string) (bool, error) {
	var v group.UserGroup
	err := db.DB.
		Where("user_id = ? AND group_id = ? AND role IN ?", userID, gid, roles).
		First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CheckGroupPermission(UID string, GID string, repos repository.UserGroupRepo) (bool, error) {
	isMember, err := HasGroupRole(UID, GID, config.GroupAccessRoles)
	if err != nil {
		return false, err
	}
	if isMember {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(context.Background(), UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %s is not a group member: permission denied", UID)
}

func CheckGroupManagePermission(UID string, GID string, repos repository.UserGroupRepo) (bool, error) {
	isManager, err := HasGroupRole(UID, GID, config.GroupUpdateRoles)
	if err != nil {
		return false, fmt.Errorf("failed to check group manager role: %w", err)
	}
	if isManager {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(context.Background(), UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %s cannot manage group %s: permission denied", UID, GID)
}

func CheckGroupAdminPermission(UID string, GID string, repos repository.UserGroupRepo) (bool, error) {
	isManager, err := HasGroupRole(UID, GID, config.GroupAdminRoles)
	if err != nil {
		return false, fmt.Errorf("failed to check group admin role: %w", err)
	}
	if isManager {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(context.Background(), UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %s cannot administer group %s: permission denied", UID, GID)
}
