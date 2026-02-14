package repository

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

type GroupRepo interface {
	Create(ctx context.Context, g *group.Group) error
	CreateGroup(ctx context.Context, g *group.Group) error
	Get(ctx context.Context, id string) (*group.Group, error)
	GetGroupByID(ctx context.Context, id string) (*group.Group, error)
	List(ctx context.Context) ([]group.Group, error)
	GetAllGroups(ctx context.Context) ([]group.Group, error)
	ListGroupsForUser(ctx context.Context, userID string) ([]group.Group, error)
	ListUsersInGroup(ctx context.Context, groupID string) ([]user.User, error)
	UpdateGroup(ctx context.Context, g *group.Group) error
	Delete(ctx context.Context, id string) error
	DeleteGroup(ctx context.Context, id string) error
	AddUser(ctx context.Context, ug *group.UserGroup) error
	RemoveUser(ctx context.Context, userID, groupID string) error
	WithTx(tx *gorm.DB) GroupRepo
}

type UserGroupRepo interface {
	IsSuperAdmin(ctx context.Context, uid string) (bool, error)
	GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error)
	CreateUserGroup(ctx context.Context, ug *group.UserGroup) error
	UpdateUserGroup(ctx context.Context, ug *group.UserGroup) error
	DeleteUserGroup(ctx context.Context, uid, gid string) error
	GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error)
	GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error)
	CountUsersByGID(ctx context.Context, gid string) (int64, error)
	WithTx(tx *gorm.DB) UserGroupRepo
}

type GroupRepoImpl struct {
	db *gorm.DB
}

type UserGroupRepoImpl struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) GroupRepo {
	return &GroupRepoImpl{db: db}
}

func NewUserGroupRepo(db *gorm.DB) UserGroupRepo {
	return &UserGroupRepoImpl{db: db}
}

func (r *GroupRepoImpl) setAliases(g *group.Group) {
	if g == nil {
		return
	}
	g.GID = g.ID
	g.GroupName = g.Name
}

func (r *GroupRepoImpl) Create(ctx context.Context, g *group.Group) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GroupRepoImpl) CreateGroup(ctx context.Context, g *group.Group) error {
	return r.Create(ctx, g)
}

func (r *GroupRepoImpl) Get(ctx context.Context, id string) (*group.Group, error) {
	var g group.Group
	err := r.db.WithContext(ctx).First(&g, "id = ?", id).Error
	r.setAliases(&g)
	return &g, err
}

func (r *GroupRepoImpl) GetGroupByID(ctx context.Context, id string) (*group.Group, error) {
	return r.Get(ctx, id)
}

func (r *GroupRepoImpl) List(ctx context.Context) ([]group.Group, error) {
	var groups []group.Group
	err := r.db.WithContext(ctx).Find(&groups).Error
	for i := range groups {
		r.setAliases(&groups[i])
	}
	return groups, err
}

func (r *GroupRepoImpl) GetAllGroups(ctx context.Context) ([]group.Group, error) {
	return r.List(ctx)
}

func (r *GroupRepoImpl) ListGroupsForUser(ctx context.Context, userID string) ([]group.Group, error) {
	var groups []group.Group
	err := r.db.WithContext(ctx).Joins("JOIN user_group ug ON ug.group_id = groups.id").
		Where("ug.user_id = ?", userID).Find(&groups).Error
	for i := range groups {
		r.setAliases(&groups[i])
	}
	return groups, err
}

func (r *GroupRepoImpl) ListUsersInGroup(ctx context.Context, groupID string) ([]user.User, error) {
	var users []user.User
	err := r.db.WithContext(ctx).Joins("JOIN user_group ug ON ug.user_id = users.id").
		Where("ug.group_id = ?", groupID).Find(&users).Error
	return users, err
}

func (r *GroupRepoImpl) UpdateGroup(ctx context.Context, g *group.Group) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *GroupRepoImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&group.Group{}, "id = ?", id).Error
}

func (r *GroupRepoImpl) DeleteGroup(ctx context.Context, id string) error {
	return r.Delete(ctx, id)
}

func (r *GroupRepoImpl) AddUser(ctx context.Context, ug *group.UserGroup) error {
	return r.db.WithContext(ctx).Create(ug).Error
}

func (r *GroupRepoImpl) RemoveUser(ctx context.Context, userID, groupID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND group_id = ?", userID, groupID).Delete(&group.UserGroup{}).Error
}

func (r *GroupRepoImpl) WithTx(tx *gorm.DB) GroupRepo {
	return &GroupRepoImpl{db: tx}
}

func (r *UserGroupRepoImpl) IsSuperAdmin(ctx context.Context, uid string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").Where("id = ? AND is_super_admin = ?", uid, true).Count(&count).Error
	return count > 0, err
}

func (r *UserGroupRepoImpl) GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
	var ug group.UserGroup
	err := r.db.WithContext(ctx).Where("user_id = ? AND group_id = ?", uid, gid).First(&ug).Error
	return &ug, err
}

func (r *UserGroupRepoImpl) CreateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	return r.db.WithContext(ctx).Create(ug).Error
}

func (r *UserGroupRepoImpl) UpdateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	return r.db.WithContext(ctx).Save(ug).Error
}

func (r *UserGroupRepoImpl) DeleteUserGroup(ctx context.Context, uid, gid string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND group_id = ?", uid, gid).Delete(&group.UserGroup{}).Error
}

func (r *UserGroupRepoImpl) GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error) {
	var ugs []group.UserGroup
	err := r.db.WithContext(ctx).Where("user_id = ?", uid).Find(&ugs).Error
	return ugs, err
}

func (r *UserGroupRepoImpl) GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error) {
	var ugs []group.UserGroup
	err := r.db.WithContext(ctx).Where("group_id = ?", gid).Find(&ugs).Error
	return ugs, err
}

func (r *UserGroupRepoImpl) CountUsersByGID(ctx context.Context, gid string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&group.UserGroup{}).Where("group_id = ?", gid).Count(&count).Error
	return count, err
}

func (r *UserGroupRepoImpl) WithTx(tx *gorm.DB) UserGroupRepo {
	return &UserGroupRepoImpl{db: tx}
}
