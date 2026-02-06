package repository

import (
	"github.com/linskybing/platform-go/internal/domain/group"
	"gorm.io/gorm"
)

type GroupRepo interface {
	GetAllGroups() ([]group.Group, error)
	GetGroupByID(id string) (group.Group, error)
	CreateGroup(group *group.Group) error
	UpdateGroup(group *group.Group) error
	DeleteGroup(id string) error
	WithTx(tx *gorm.DB) GroupRepo
}

type DBGroupRepo struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *DBGroupRepo {
	return &DBGroupRepo{
		db: db,
	}
}

func (r *DBGroupRepo) GetAllGroups() ([]group.Group, error) {
	var groups []group.Group
	err := r.db.Find(&groups).Error
	return groups, err
}

func (r *DBGroupRepo) GetGroupByID(id string) (group.Group, error) {
	var group group.Group
	err := r.db.First(&group, "g_id = ?", id).Error
	return group, err
}

func (r *DBGroupRepo) CreateGroup(group *group.Group) error {
	return r.db.Create(group).Error
}

func (r *DBGroupRepo) UpdateGroup(group *group.Group) error {
	return r.db.Save(group).Error
}

func (r *DBGroupRepo) DeleteGroup(id string) error {
	return r.db.Delete(&group.Group{}, "g_id = ?", id).Error
}

func (r *DBGroupRepo) WithTx(tx *gorm.DB) GroupRepo {
	if tx == nil {
		return r
	}
	return &DBGroupRepo{
		db: tx,
	}
}
