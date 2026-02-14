package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubUserGroupRepoUGWithFuncs struct {
	createUserGroup func(ctx context.Context, ug *group.UserGroup) error
	updateUserGroup func(ctx context.Context, ug *group.UserGroup) error
	deleteUserGroup func(ctx context.Context, uid, gid string) error
	getUserGroup    func(ctx context.Context, uid, gid string) (*group.UserGroup, error)
}

func (s *stubUserGroupRepoUGWithFuncs) IsSuperAdmin(ctx context.Context, uid string) (bool, error) {
	return false, nil
}

func (s *stubUserGroupRepoUGWithFuncs) GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
	if s.getUserGroup != nil {
		return s.getUserGroup(ctx, uid, gid)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubUserGroupRepoUGWithFuncs) CreateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	if s.createUserGroup != nil {
		return s.createUserGroup(ctx, ug)
	}
	return nil
}

func (s *stubUserGroupRepoUGWithFuncs) UpdateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	if s.updateUserGroup != nil {
		return s.updateUserGroup(ctx, ug)
	}
	return nil
}

func (s *stubUserGroupRepoUGWithFuncs) DeleteUserGroup(ctx context.Context, uid, gid string) error {
	if s.deleteUserGroup != nil {
		return s.deleteUserGroup(ctx, uid, gid)
	}
	return nil
}

func (s *stubUserGroupRepoUGWithFuncs) GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (s *stubUserGroupRepoUGWithFuncs) GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (s *stubUserGroupRepoUGWithFuncs) CountUsersByGID(ctx context.Context, gid string) (int64, error) {
	return 0, nil
}

func (s *stubUserGroupRepoUGWithFuncs) WithTx(tx *gorm.DB) repository.UserGroupRepo { return s }
