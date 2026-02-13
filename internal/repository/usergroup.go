package repository

import (
	"errors"

	"github.com/linskybing/platform-go/internal/domain/group"
	"gorm.io/gorm"
)

type UserGroupRepo interface {
	CreateUserGroup(userGroup *group.UserGroup) error
	UpdateUserGroup(userGroup *group.UserGroup) error
	DeleteUserGroup(uid, gid string) error
	GetUserGroupsByUID(uid string) ([]group.UserGroup, error)
	GetUserGroupsByGID(gid string) ([]group.UserGroup, error)
	GetUserGroup(uid, gid string) (group.UserGroup, error)
	CountUsersByGID(gid string) (int64, error)
	IsSuperAdmin(uid string) (bool, error)
	GetUserRoleInGroup(uid string, gid string) (string, error)
	WithTx(tx *gorm.DB) UserGroupRepo
}

type DBUserGroupRepo struct {
	db *gorm.DB
}

func NewUserGroupRepo(db *gorm.DB) *DBUserGroupRepo {
	return &DBUserGroupRepo{
		db: db,
	}
}

func (r *DBUserGroupRepo) CreateUserGroup(userGroup *group.UserGroup) error {
	return r.db.Create(userGroup).Error
}

func (r *DBUserGroupRepo) UpdateUserGroup(userGroup *group.UserGroup) error {
	return r.db.Save(userGroup).Error
}

func (r *DBUserGroupRepo) DeleteUserGroup(uid, gid string) error {
	return r.db.Where("u_id = ? AND g_id = ?", uid, gid).Delete(&group.UserGroup{}).Error
}

func (r *DBUserGroupRepo) GetUserGroupsByUID(uid string) ([]group.UserGroup, error) {
	var userGroups []group.UserGroup
	err := r.db.
		Where("u_id = ?", uid).
		Preload("Group").
		Find(&userGroups).Error
	return userGroups, err
}

func (r *DBUserGroupRepo) GetUserGroupsByGID(gid string) ([]group.UserGroup, error) {
	var userGroups []group.UserGroup
	err := r.db.
		Where("g_id = ?", gid).
		Preload("User").
		Find(&userGroups).Error
	return userGroups, err
}

func (r *DBUserGroupRepo) GetUserGroup(uid, gid string) (group.UserGroup, error) {
	var userGroup group.UserGroup
	err := r.db.First(&userGroup, "u_id = ? AND g_id = ?", uid, gid).Error
	return userGroup, err
}

func (r *DBUserGroupRepo) CountUsersByGID(gid string) (int64, error) {
	var count int64
	err := r.db.Model(&group.UserGroup{}).Where("g_id = ?", gid).Count(&count).Error
	return count, err
}

func (r *DBUserGroupRepo) IsSuperAdmin(uid string) (bool, error) {
	var count int64
	err := r.db.Table("user_group").
		Joins("JOIN group_list g ON g.g_id = user_group.g_id").
		Where("user_group.u_id = ? AND g.group_name = ? AND user_group.role = ?", uid, "super", "admin").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DBUserGroupRepo) GetUserRoleInGroup(uid string, gid string) (string, error) {
	var role string
	err := r.db.Table("user_group").
		Select("role").
		Where("u_id = ? AND g_id = ?", uid, gid).
		Limit(1).
		Scan(&role).Error

	if err == nil && role != "" {
		return role, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || role == "" {
		return "", gorm.ErrRecordNotFound
	}

	return role, err
}

func (r *DBUserGroupRepo) WithTx(tx *gorm.DB) UserGroupRepo {
	if tx == nil {
		return r
	}
	return &DBUserGroupRepo{
		db: tx,
	}
}
