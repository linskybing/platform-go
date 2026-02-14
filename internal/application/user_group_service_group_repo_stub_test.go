package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubGroupRepoLite struct {
	getGroupByID func(ctx context.Context, id string) (*group.Group, error)
}

func (s *stubGroupRepoLite) Create(ctx context.Context, g *group.Group) error      { return nil }
func (s *stubGroupRepoLite) CreateGroup(ctx context.Context, g *group.Group) error { return nil }
func (s *stubGroupRepoLite) Get(ctx context.Context, id string) (*group.Group, error) {
	return s.GetGroupByID(ctx, id)
}
func (s *stubGroupRepoLite) GetGroupByID(ctx context.Context, id string) (*group.Group, error) {
	if s.getGroupByID != nil {
		return s.getGroupByID(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubGroupRepoLite) List(ctx context.Context) ([]group.Group, error) { return nil, nil }
func (s *stubGroupRepoLite) GetAllGroups(ctx context.Context) ([]group.Group, error) {
	return nil, nil
}
func (s *stubGroupRepoLite) ListGroupsForUser(ctx context.Context, userID string) ([]group.Group, error) {
	return nil, nil
}
func (s *stubGroupRepoLite) ListUsersInGroup(ctx context.Context, groupID string) ([]user.User, error) {
	return nil, nil
}
func (s *stubGroupRepoLite) UpdateGroup(ctx context.Context, g *group.Group) error { return nil }
func (s *stubGroupRepoLite) Delete(ctx context.Context, id string) error           { return nil }
func (s *stubGroupRepoLite) DeleteGroup(ctx context.Context, id string) error      { return nil }
func (s *stubGroupRepoLite) AddUser(ctx context.Context, ug *group.UserGroup) error {
	return nil
}
func (s *stubGroupRepoLite) RemoveUser(ctx context.Context, userID, groupID string) error {
	return nil
}
func (s *stubGroupRepoLite) WithTx(tx *gorm.DB) repository.GroupRepo { return s }
