package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubUserGroupRepoUG struct {
	createUserGroup func(ctx context.Context, ug *group.UserGroup) error
	updateUserGroup func(ctx context.Context, ug *group.UserGroup) error
	deleteUserGroup func(ctx context.Context, uid, gid string) error
	getUserGroup    func(ctx context.Context, uid, gid string) (*group.UserGroup, error)
}

func (s *stubUserGroupRepoUG) IsSuperAdmin(ctx context.Context, uid string) (bool, error) {
	return false, nil
}

func (s *stubUserGroupRepoUG) GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
	if s.getUserGroup != nil {
		return s.getUserGroup(ctx, uid, gid)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubUserGroupRepoUG) CreateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	if s.createUserGroup != nil {
		return s.createUserGroup(ctx, ug)
	}
	return nil
}

func (s *stubUserGroupRepoUG) UpdateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	if s.updateUserGroup != nil {
		return s.updateUserGroup(ctx, ug)
	}
	return nil
}

func (s *stubUserGroupRepoUG) DeleteUserGroup(ctx context.Context, uid, gid string) error {
	if s.deleteUserGroup != nil {
		return s.deleteUserGroup(ctx, uid, gid)
	}
	return nil
}

func (s *stubUserGroupRepoUG) GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (s *stubUserGroupRepoUG) GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (s *stubUserGroupRepoUG) CountUsersByGID(ctx context.Context, gid string) (int64, error) {
	return 0, nil
}

func (s *stubUserGroupRepoUG) WithTx(tx *gorm.DB) repository.UserGroupRepo { return s }
