package repositories

import (
	"errors"

	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/models"
	"gorm.io/gorm"
)

func GetAllProjectGroupViews() ([]models.ProjectGroupView, error) {
	var results []models.ProjectGroupView
	err := db.DB.Find(&results).Error
	return results, err
}

func GetProjectResourcesByGroupID(groupID uint) ([]models.ProjectResourceView, error) {
	var results []models.ProjectResourceView
	err := db.DB.Where("g_id = ?", groupID).Find(&results).Error
	return results, err
}

func GetGroupResourcesByGroupID(groupID uint) ([]models.GroupResourceView, error) {
	var results []models.GroupResourceView
	err := db.DB.Where("g_id = ?", groupID).Find(&results).Error
	return results, err
}

func GetGroupIDByResourceID(rID uint) (uint, error) {
	type result struct {
		GID uint `gorm:"column:g_id"`
	}

	var res result

	err := db.DB.Select("p.g_id").Joins("JOIN project p ON r.p_id = p.p_id").Where("r.r_id = ?", rID).
		Take(&res).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, gorm.ErrRecordNotFound
	}
	if err != nil {
		return 0, err
	}

	return res.GID, nil
}
