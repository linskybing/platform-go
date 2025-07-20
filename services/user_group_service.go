package services

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/utils"
)

func CreateUserGroup(c *gin.Context, userGroup *models.UserGroup) (*models.UserGroup, error) {
	if err := repositories.CreateUserGroup(userGroup); err != nil {
		return nil, err
	}

	utils.LogAuditWithConsole(c, "create", "user_group",
		fmt.Sprintf("u_id=%d,g_id=%d", userGroup.UID, userGroup.GID),
		nil, *userGroup, "")

	return userGroup, nil
}

func UpdateUserGroup(c *gin.Context, userGroup *models.UserGroup) (*models.UserGroup, error) {
	oldUserGroup, err := repositories.GetUserGroup(userGroup.UID, userGroup.GID)
	if err != nil {
		return nil, err
	}

	if err := repositories.UpdateUserGroup(userGroup); err != nil {
		return nil, err
	}

	utils.LogAuditWithConsole(c, "update", "user_group",
		fmt.Sprintf("u_id=%d,g_id=%d", userGroup.UID, userGroup.GID),
		oldUserGroup, *userGroup, "")

	return userGroup, nil
}

func DeleteUserGroup(c *gin.Context, uid, gid uint) error {
	oldUserGroup, err := repositories.GetUserGroup(uid, gid)
	if err != nil {
		return err
	}

	if err := repositories.DeleteUserGroup(uid, gid); err != nil {
		return err
	}

	utils.LogAuditWithConsole(c, "delete", "user_group",
		fmt.Sprintf("u_id=%d,g_id=%d", uid, gid),
		oldUserGroup, nil, "")

	return nil
}

func GetUserGroup(uid, gid uint) (models.UserGroupView, error) {
	return repositories.GetUserGroup(uid, gid)
}

func GetUserGroupsByUID(uid uint) ([]models.UserGroupView, error) {
	return repositories.GetUserGroupsByUID(uid)
}

func GetUserGroupsByGID(gid uint) ([]models.UserGroupView, error) {
	return repositories.GetUserGroupsByGID(gid)
}
