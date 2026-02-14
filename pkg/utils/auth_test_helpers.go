package utils

import (
	"context"
	"errors"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type mockUserGroupRepo struct {
	roles       map[string]map[string][]string
	superAdmins map[string]bool
	shouldErr   bool
}

func (m *mockUserGroupRepo) IsSuperAdmin(ctx context.Context, uid string) (bool, error) {
	if m.shouldErr {
		return false, errors.New("database error")
	}
	return m.superAdmins[uid], nil
}

func (m *mockUserGroupRepo) GetRoles(uid, gid string) ([]string, error) {
	if m.shouldErr {
		return nil, errors.New("database error")
	}
	if m.roles[uid] != nil {
		return m.roles[uid][gid], nil
	}
	return []string{}, nil
}

func (m *mockUserGroupRepo) CreateUserGroup(ctx context.Context, userGroup *group.UserGroup) error {
	return nil
}

func (m *mockUserGroupRepo) UpdateUserGroup(ctx context.Context, userGroup *group.UserGroup) error {
	return nil
}

func (m *mockUserGroupRepo) DeleteUserGroup(ctx context.Context, uid, gid string) error {
	return nil
}

func (m *mockUserGroupRepo) GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (m *mockUserGroupRepo) GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error) {
	return nil, nil
}

func (m *mockUserGroupRepo) GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
	return &group.UserGroup{}, nil
}

func (m *mockUserGroupRepo) CountUsersByGID(ctx context.Context, gid string) (int64, error) {
	return 0, nil
}

func (m *mockUserGroupRepo) WithTx(tx *gorm.DB) repository.UserGroupRepo {
	return m
}
