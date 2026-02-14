package project

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/gorm"
)

type stubProjectRepo struct {
	getNodeByOwner func(ctx context.Context, ownerID string) (*project.Project, error)
	getNode        func(ctx context.Context, id string) (*project.Project, error)
	createNode     func(ctx context.Context, node *project.Project) error
	createPlan     func(ctx context.Context, plan *project.ResourcePlan) error
	getProjectByID func(ctx context.Context, id string) (*project.Project, error)
}

func (s *stubProjectRepo) CreateProject(ctx context.Context, node *project.Project) error {
	return s.CreateNode(ctx, node)
}
func (s *stubProjectRepo) CreateNode(ctx context.Context, node *project.Project) error {
	if s.createNode != nil {
		return s.createNode(ctx, node)
	}
	return nil
}
func (s *stubProjectRepo) UpdateProject(ctx context.Context, node *project.Project) error { return nil }
func (s *stubProjectRepo) UpdateNode(ctx context.Context, node *project.Project) error    { return nil }
func (s *stubProjectRepo) GetProjectByID(ctx context.Context, id string) (*project.Project, error) {
	if s.getProjectByID != nil {
		return s.getProjectByID(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) GetNode(ctx context.Context, id string) (*project.Project, error) {
	if s.getNode != nil {
		return s.getNode(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error) {
	if s.getNodeByOwner != nil {
		return s.getNodeByOwner(ctx, ownerID)
	}
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
	if s.createPlan != nil {
		return s.createPlan(ctx, plan)
	}
	return nil
}
func (s *stubProjectRepo) UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return nil
}
func (s *stubProjectRepo) GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubProjectRepo) WithTx(tx *gorm.DB) repository.ProjectRepo { return s }

func setupProjectService(t *testing.T) (*ProjectService, *stubProjectRepo, *gin.Context) {
	t.Helper()
	stub := &stubProjectRepo{}
	repos := &repository.Repos{Project: stub}
	svc := NewProjectService(repos, nil)
	ctx, _ := gin.CreateTestContext(nil)
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}
	return svc, stub, ctx
}
