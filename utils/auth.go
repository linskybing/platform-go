package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/types"
	"gorm.io/gorm"
)

func IsSuperAdmin(uid uint) (bool, error) {
	var view models.UserGroupView
	err := db.DB.
		Where("u_id = ? AND group_name = ? AND role = ?", uid, "super", "admin").
		First(&view).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetUserIDFromContext(c *gin.Context) (uint, error) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return 0, errors.New("user claims not found in context")
	}

	claims, ok := claimsVal.(*types.Claims)
	if !ok {
		return 0, errors.New("invalid user claims type")
	}

	return claims.UserID, nil
}
