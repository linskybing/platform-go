package project

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/project"
)

// TestCreateProject verifies the successful creation and persistence of a project.
func TestCreateProject(t *testing.T) {
	svc, projRepo, c := setupProjectService(t)
	t.Run("success", func(t *testing.T) {
		dto := project.CreateProjectDTO{ProjectName: "p1", GID: "g1"}
		projRepo.getNodeByOwner = func(ctx context.Context, ownerID string) (*project.Project, error) {
			return &project.Project{ID: "g1"}, nil
		}
		projRepo.createNode = func(ctx context.Context, p *project.Project) error {
			p.ID = "1"
			return nil
		}
		res, err := svc.CreateProject(c, dto)
		if err != nil || res.ID != "1" {
			t.Fail()
		}
	})
	t.Run("group not found", func(t *testing.T) {
		dto := project.CreateProjectDTO{ProjectName: "p1", GID: "g2"}
		projRepo.getNodeByOwner = func(ctx context.Context, ownerID string) (*project.Project, error) {
			return nil, errors.New("not found")
		}
		projRepo.getNode = func(ctx context.Context, id string) (*project.Project, error) {
			return nil, errors.New("not found")
		}
		_, err := svc.CreateProject(c, dto)
		if err == nil {
			t.Fail()
		}
	})
}

// TestGetProject checks the retrieval of project details by ID.
func TestGetProject(t *testing.T) {
	svc, projRepo, _ := setupProjectService(t)
	t.Run("success", func(t *testing.T) {
		projRepo.getProjectByID = func(ctx context.Context, id string) (*project.Project, error) {
			return &project.Project{PID: "1", ProjectName: "p1"}, nil
		}
		res, err := svc.GetProject("1")
		if err != nil || res.ProjectName != "p1" {
			t.Fail()
		}
	})
}
