package utils

import (
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

func IsSuperAdmin(uid uint, repos repository.UserGroupRepo) (bool, error) {
	if uid == 1 {
		return true, nil
	}
	return repos.IsSuperAdmin(uid)
}

var GetUserIDFromContext = func(c *gin.Context) (uint, error) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return 0, fmt.Errorf("user claims not found in context: missing claims")
	}

	claims, ok := claimsVal.(*types.Claims)
	if !ok {
		return 0, fmt.Errorf("invalid user claims type (expected *types.Claims): type assertion failed")
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

func HasGroupRole(userID uint, gid uint, roles []string) (bool, error) {
	var v group.UserGroup
	err := db.DB.
		Where("u_id = ? AND g_id = ? AND role IN ?", userID, gid, roles).
		First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var ug group.UserGroup
			tableErr := db.DB.Where("u_id = ? AND g_id = ? AND role IN ?", userID, gid, roles).First(&ug).Error
			if tableErr != nil {
				if errors.Is(tableErr, gorm.ErrRecordNotFound) {
					return false, nil
				}
				return false, tableErr
			}
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func CheckGroupPermission(UID uint, GID uint, repos repository.UserGroupRepo) (bool, error) {
	isMember, err := HasGroupRole(UID, GID, config.GroupAccessRoles)
	if err != nil {
		return false, err
	}
	if isMember {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %d is not a group member: permission denied", UID)
}

func CheckGroupManagePermission(UID uint, GID uint, repos repository.UserGroupRepo) (bool, error) {
	isManager, err := HasGroupRole(UID, GID, config.GroupUpdateRoles)
	if err != nil {
		return false, fmt.Errorf("failed to check group manager role: %w", err)
	}
	if isManager {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %d cannot manage group %d: permission denied", UID, GID)
}

func CheckGroupAdminPermission(UID uint, GID uint, repos repository.UserGroupRepo) (bool, error) {
	isManager, err := HasGroupRole(UID, GID, config.GroupAdminRoles)
	if err != nil {
		return false, fmt.Errorf("failed to check group admin role: %w", err)
	}
	if isManager {
		return true, nil
	}

	isSuper, err := repos.IsSuperAdmin(UID)
	if err != nil {
		return false, fmt.Errorf("failed to check super admin status: %w", err)
	}
	if isSuper {
		return true, nil
	}

	return false, fmt.Errorf("user %d cannot administer group %d: permission denied", UID, GID)
}
