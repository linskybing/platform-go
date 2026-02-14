package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubProjectRepoLiteUserGroup struct {
	listProjectsByGroup func(ctx context.Context, gid string) ([]project.Project, error)
}

func (s *stubProjectRepoLiteUserGroup) CreateProject(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) CreateNode(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) UpdateProject(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) UpdateNode(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) GetProjectByID(ctx context.Context, id string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLiteUserGroup) GetNode(ctx context.Context, id string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLiteUserGroup) GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLiteUserGroup) ListProjects(ctx context.Context) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) ListNodes(ctx context.Context, parentID *string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) ListProjectsByGroup(ctx context.Context, gid string) ([]project.Project, error) {
	if s.listProjectsByGroup != nil {
		return s.listProjectsByGroup(ctx, gid)
	}
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) ListDescendantProjects(ctx context.Context, groupOwnerIDs []string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) GetSubtree(ctx context.Context, rootID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLiteUserGroup) MoveNode(ctx context.Context, nodeID, newParentID string) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) DeleteNode(ctx context.Context, id string) error { return nil }
func (s *stubProjectRepoLiteUserGroup) CreateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepoLiteUserGroup) GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLiteUserGroup) WithTx(tx *gorm.DB) repository.ProjectRepo { return s }
