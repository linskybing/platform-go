package application_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/gorm"
)

type stubGroupRepo struct {
	list       func(ctx context.Context) ([]group.Group, error)
	get        func(ctx context.Context, id string) (*group.Group, error)
	create     func(ctx context.Context, g *group.Group) error
	update     func(ctx context.Context, g *group.Group) error
	deleteFunc func(ctx context.Context, id string) error
}

func (s *stubGroupRepo) Create(ctx context.Context, g *group.Group) error {
	if s.create != nil {
		return s.create(ctx, g)
	}
	return nil
}

func (s *stubGroupRepo) CreateGroup(ctx context.Context, g *group.Group) error {
	return s.Create(ctx, g)
}

func (s *stubGroupRepo) Get(ctx context.Context, id string) (*group.Group, error) {
	if s.get != nil {
		return s.get(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubGroupRepo) GetGroupByID(ctx context.Context, id string) (*group.Group, error) {
	return s.Get(ctx, id)
}

func (s *stubGroupRepo) List(ctx context.Context) ([]group.Group, error) {
	if s.list != nil {
		return s.list(ctx)
	}
	return nil, nil
}

func (s *stubGroupRepo) GetAllGroups(ctx context.Context) ([]group.Group, error) { return s.List(ctx) }

func (s *stubGroupRepo) ListGroupsForUser(ctx context.Context, userID string) ([]group.Group, error) {
	return nil, nil
}

func (s *stubGroupRepo) ListUsersInGroup(ctx context.Context, groupID string) ([]user.User, error) {
	return nil, nil
}

func (s *stubGroupRepo) UpdateGroup(ctx context.Context, g *group.Group) error {
	if s.update != nil {
		return s.update(ctx, g)
	}
	return nil
}

func (s *stubGroupRepo) Delete(ctx context.Context, id string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, id)
	}
	return nil
}

func (s *stubGroupRepo) DeleteGroup(ctx context.Context, id string) error { return s.Delete(ctx, id) }

func (s *stubGroupRepo) AddUser(ctx context.Context, ug *group.UserGroup) error { return nil }
func (s *stubGroupRepo) RemoveUser(ctx context.Context, userID, groupID string) error {
	return nil
}

func (s *stubGroupRepo) WithTx(tx *gorm.DB) repository.GroupRepo { return s }

func setupGroupService(t *testing.T) (*application.GroupService, *stubGroupRepo, *gin.Context) {
	t.Helper()
	stubGroup := &stubGroupRepo{}
	repos := &repository.Repos{Group: stubGroup}
	svc := application.NewGroupService(repos)
	ctx, _ := gin.CreateTestContext(nil)
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}
	return svc, stubGroup, ctx
}
