package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubProjectRepo struct {
	getProjectByID func(ctx context.Context, id string) (*project.Project, error)
}

func (s *stubProjectRepo) CreateProject(ctx context.Context, node *project.Project) error { return nil }
func (s *stubProjectRepo) CreateNode(ctx context.Context, node *project.Project) error    { return nil }
func (s *stubProjectRepo) UpdateProject(ctx context.Context, node *project.Project) error { return nil }
func (s *stubProjectRepo) UpdateNode(ctx context.Context, node *project.Project) error    { return nil }
func (s *stubProjectRepo) GetProjectByID(ctx context.Context, id string) (*project.Project, error) {
	if s.getProjectByID != nil {
		return s.getProjectByID(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) GetNode(ctx context.Context, id string) (*project.Project, error) {
	return s.GetProjectByID(ctx, id)
}
func (s *stubProjectRepo) GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) ListProjects(ctx context.Context) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) ListNodes(ctx context.Context, parentID *string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) ListProjectsByGroup(ctx context.Context, gid string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) ListDescendantProjects(ctx context.Context, groupOwnerIDs []string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) GetSubtree(ctx context.Context, rootID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepo) MoveNode(ctx context.Context, nodeID, newParentID string) error { return nil }
func (s *stubProjectRepo) DeleteNode(ctx context.Context, id string) error                { return nil }
func (s *stubProjectRepo) CreateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepo) UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepo) GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) WithTx(tx *gorm.DB) repository.ProjectRepo { return s }

type stubUserGroupRepo struct {
	getUserGroup func(ctx context.Context, uid, gid string) (*group.UserGroup, error)
}

func (s *stubUserGroupRepo) IsSuperAdmin(ctx context.Context, uid string) (bool, error) {
	return false, nil
}
func (s *stubUserGroupRepo) GetUserGroup(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
	if s.getUserGroup != nil {
		return s.getUserGroup(ctx, uid, gid)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserGroupRepo) CreateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	return nil
}
func (s *stubUserGroupRepo) UpdateUserGroup(ctx context.Context, ug *group.UserGroup) error {
	return nil
}
func (s *stubUserGroupRepo) DeleteUserGroup(ctx context.Context, uid, gid string) error { return nil }
func (s *stubUserGroupRepo) GetUserGroupsByUID(ctx context.Context, uid string) ([]group.UserGroup, error) {
	return nil, nil
}
func (s *stubUserGroupRepo) GetUserGroupsByGID(ctx context.Context, gid string) ([]group.UserGroup, error) {
	return nil, nil
}
func (s *stubUserGroupRepo) CountUsersByGID(ctx context.Context, gid string) (int64, error) {
	return 0, nil
}
func (s *stubUserGroupRepo) WithTx(tx *gorm.DB) repository.UserGroupRepo { return s }

type stubUserRepo struct {
	listUsersByProjectID func(ctx context.Context, pid string) ([]user.User, error)
}

func (s *stubUserRepo) Create(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepo) Get(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) GetUserRawByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) GetUsernameByID(ctx context.Context, id string) (string, error) {
	return "", gorm.ErrRecordNotFound
}
func (s *stubUserRepo) List(ctx context.Context) ([]user.User, error)        { return nil, nil }
func (s *stubUserRepo) GetAllUsers(ctx context.Context) ([]user.User, error) { return nil, nil }
func (s *stubUserRepo) ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepo) ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error) {
	if s.listUsersByProjectID != nil {
		return s.listUsersByProjectID(ctx, pid)
	}
	return nil, nil
}
func (s *stubUserRepo) SaveUser(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepo) Delete(ctx context.Context, id string) error      { return nil }
func (s *stubUserRepo) WithTx(tx *gorm.DB) repository.UserRepo           { return s }
