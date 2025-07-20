package repositories

import (
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/models"
)

func CreateUserGroup(userGroup *models.UserGroup) error {
	return db.DB.Create(userGroup).Error
}

func UpdateUserGroup(userGroup *models.UserGroup) error {
	return db.DB.Save(userGroup).Error
}

func DeleteUserGroup(uid, gid uint) error {
	return db.DB.Where("u_id = ? AND g_id = ?", uid, gid).Delete(&models.UserGroup{}).Error
}

func GetUserGroupsByUID(uid uint) ([]models.UserGroupView, error) {
	var userGroups []models.UserGroupView
	err := db.DB.
		Where("u_id = ?", uid).
		Find(&userGroups).Error
	return userGroups, err
}

func GetUserGroupsByGID(gid uint) ([]models.UserGroupView, error) {
	var userGroups []models.UserGroupView
	err := db.DB.
		Where("g_id = ?", gid).
		Find(&userGroups).Error
	return userGroups, err
}

func GetUserGroup(uid, gid uint) (models.UserGroupView, error) {
	var userGroup models.UserGroupView
	err := db.DB.First(&userGroup, "u_id = ? AND g_id = ?", uid, gid).Error
	return userGroup, err
}
