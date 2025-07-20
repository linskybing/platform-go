package services

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/utils"
)

var ErrReservedGroupName = errors.New("cannot use reserved group name 'super'")

func ListGroups() ([]models.Group, error) {
	return repositories.GetAllGroups()
}

func GetGroup(id uint) (models.Group, error) {
	return repositories.GetGroupByID(id)
}

func CreateGroup(c *gin.Context, input dto.GroupCreateDTO) (models.Group, error) {
	if input.GroupName == "super" {
		return models.Group{}, ErrReservedGroupName
	}

	group := models.Group{
		GroupName: input.GroupName,
	}
	if input.Description != nil {
		group.Description = *input.Description
	}

	err := repositories.CreateGroup(&group)
	if err != nil {
		return models.Group{}, err
	}
	utils.LogAuditWithConsole(c, "create", "group", fmt.Sprintf("g_id=%d", group.GID), nil, group, "")

	return group, nil
}

func UpdateGroup(c *gin.Context, id uint, input dto.GroupUpdateDTO) (models.Group, error) {
	group, err := repositories.GetGroupByID(id)
	if err != nil {
		return models.Group{}, err
	}

	oldGroup := group

	if input.GroupName != nil {
		if *input.GroupName == "super" {
			return models.Group{}, ErrReservedGroupName
		}
		group.GroupName = *input.GroupName
	}
	if input.Description != nil {
		group.Description = *input.Description
	}

	err = repositories.UpdateGroup(&group)
	if err != nil {
		return models.Group{}, err
	}

	utils.LogAuditWithConsole(c, "update", "group", fmt.Sprintf("g_id=%d", group.GID), oldGroup, group, "")

	return group, nil
}

func DeleteGroup(c *gin.Context, id uint) error {
	group, err := repositories.GetGroupByID(id)
	if err != nil {
		return err
	}

	err = repositories.DeleteGroup(id)
	if err != nil {
		return err
	}

	utils.LogAuditWithConsole(c, "delete", "group", fmt.Sprintf("g_id=%d", group.GID), group, nil, "")

	return nil
}
