package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubProjectRepoLite struct {
	listProjectsByGroup func(ctx context.Context, gid string) ([]project.Project, error)
}

func (s *stubProjectRepoLite) CreateProject(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLite) CreateNode(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLite) UpdateProject(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLite) UpdateNode(ctx context.Context, node *project.Project) error {
	return nil
}
func (s *stubProjectRepoLite) GetProjectByID(ctx context.Context, id string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLite) GetNode(ctx context.Context, id string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLite) GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLite) ListProjects(ctx context.Context) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLite) ListNodes(ctx context.Context, parentID *string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLite) ListProjectsByGroup(ctx context.Context, gid string) ([]project.Project, error) {
	if s.listProjectsByGroup != nil {
		return s.listProjectsByGroup(ctx, gid)
	}
	return nil, nil
}
func (s *stubProjectRepoLite) ListDescendantProjects(ctx context.Context, groupOwnerIDs []string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLite) GetSubtree(ctx context.Context, rootID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLite) GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error) {
	return nil, nil
}
func (s *stubProjectRepoLite) MoveNode(ctx context.Context, nodeID, newParentID string) error {
	return nil
}
func (s *stubProjectRepoLite) DeleteNode(ctx context.Context, id string) error { return nil }
func (s *stubProjectRepoLite) CreateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepoLite) UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepoLite) GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepoLite) WithTx(tx *gorm.DB) repository.ProjectRepo { return s }
