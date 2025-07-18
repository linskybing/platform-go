package repositories

import (
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/models"
)

func GetAllGroups() ([]models.Group, error) {
	var groups []models.Group
	err := db.DB.Find(&groups).Error
	return groups, err
}

func GetGroupByID(id uint) (models.Group, error) {
	var group models.Group
	err := db.DB.First(&group, id).Error
	return group, err
}

func CreateGroup(group *models.Group) error {
	return db.DB.Create(group).Error
}

func UpdateGroup(group *models.Group) error {
	return db.DB.Save(group).Error
}

func DeleteGroup(id uint) error {
	return db.DB.Delete(&models.Group{}, id).Error
}
