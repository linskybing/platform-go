package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubGroupRepoLiteAlt struct {
	getGroupByID func(ctx context.Context, id string) (*group.Group, error)
}

func (s *stubGroupRepoLiteAlt) Create(ctx context.Context, g *group.Group) error      { return nil }
func (s *stubGroupRepoLiteAlt) CreateGroup(ctx context.Context, g *group.Group) error { return nil }
func (s *stubGroupRepoLiteAlt) Get(ctx context.Context, id string) (*group.Group, error) {
	return s.GetGroupByID(ctx, id)
}
func (s *stubGroupRepoLiteAlt) GetGroupByID(ctx context.Context, id string) (*group.Group, error) {
	if s.getGroupByID != nil {
		return s.getGroupByID(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubGroupRepoLiteAlt) List(ctx context.Context) ([]group.Group, error) { return nil, nil }
func (s *stubGroupRepoLiteAlt) GetAllGroups(ctx context.Context) ([]group.Group, error) {
	return nil, nil
}
func (s *stubGroupRepoLiteAlt) ListGroupsForUser(ctx context.Context, userID string) ([]group.Group, error) {
	return nil, nil
}
func (s *stubGroupRepoLiteAlt) ListUsersInGroup(ctx context.Context, groupID string) ([]user.User, error) {
	return nil, nil
}
func (s *stubGroupRepoLiteAlt) UpdateGroup(ctx context.Context, g *group.Group) error { return nil }
func (s *stubGroupRepoLiteAlt) Delete(ctx context.Context, id string) error           { return nil }
func (s *stubGroupRepoLiteAlt) DeleteGroup(ctx context.Context, id string) error      { return nil }
func (s *stubGroupRepoLiteAlt) AddUser(ctx context.Context, ug *group.UserGroup) error {
	return nil
}
func (s *stubGroupRepoLiteAlt) RemoveUser(ctx context.Context, userID, groupID string) error {
	return nil
}
func (s *stubGroupRepoLiteAlt) WithTx(tx *gorm.DB) repository.GroupRepo { return s }
