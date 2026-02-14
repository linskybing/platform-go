package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	domProject "github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/gorm"
)

// CreateProject handles the creation of a new project node and its resource plan.
func (s *ProjectService) CreateProject(c *gin.Context, input domProject.CreateProjectDTO) (*domProject.Project, error) {
	ctx := context.Background()
	groupNode, err := s.Repos.Project.GetNodeByOwner(ctx, input.GID)
	if err != nil {
		groupNode, err = s.Repos.Project.GetNode(ctx, input.GID)
		if err != nil {
			return nil, fmt.Errorf("parent group node not found for GID %s: %w", input.GID, err)
		}
	}

	p := &domProject.Project{
		Name:        input.ProjectName,
		Description: getValue(input.Description),
		ParentID:    &groupNode.ID,
		OwnerID:     &input.GID,
	}

	if err := s.Repos.Project.CreateNode(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create project node: %w", err)
	}

	plan := &domProject.ResourcePlan{
		ProjectID:  p.ID,
		GPULimit:   getValueInt(input.GPUQuota),
		WeekWindow: "[0,604800)", // Default: Full week
	}
	if err := s.Repos.Project.CreateResourcePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create resource plan: %w", err)
	}
	p.ResourcePlan = *plan

	s.invalidateProjectCache(p.ID)
	s.logAudit(c, "create", p.ID, nil, p)
	return p, nil
}

// UpdateProject modifies an existing project's metadata or resource limits.
func (s *ProjectService) UpdateProject(c *gin.Context, id string, input domProject.UpdateProjectDTO) (*domProject.Project, error) {
	ctx := context.Background()
	p, err := s.Repos.Project.GetNode(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("project not found: %w", err)
	}
	old := *p

	if input.ProjectName != nil {
		p.Name = *input.ProjectName
	}
	if input.Description != nil {
		p.Description = *input.Description
	}

	if input.GID != nil {
		newGroupNode, err := s.Repos.Project.GetNodeByOwner(ctx, *input.GID)
		if err == nil {
			if err := s.Repos.Project.MoveNode(ctx, p.ID, newGroupNode.ID); err != nil {
				return nil, err
			}
			p.ParentID = &newGroupNode.ID
			p.OwnerID = input.GID // Update OwnerID as well
		}
	}

	if err := s.Repos.Project.UpdateNode(ctx, p); err != nil {
		return nil, err
	}

	if input.GPUQuota != nil {
		p.ResourcePlan.GPULimit = *input.GPUQuota
		if err := s.Repos.Project.UpdateResourcePlan(ctx, &p.ResourcePlan); err != nil {
			return nil, err
		}
	}

	s.invalidateProjectCache(p.ID)
	s.logAudit(c, "update", p.ID, old, p)
	return p, nil
}

// DeleteProject removes a project and its associated cloud resources.
func (s *ProjectService) DeleteProject(c *gin.Context, id string) error {
	ctx := context.Background()
	p, err := s.Repos.Project.GetNode(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("project not found")
	}
	if err := s.RemoveProjectResources(id); err != nil {
		return err
	}
	if err := s.Repos.Project.DeleteNode(ctx, id); err != nil {
		return err
	}
	s.invalidateProjectCache(id)
	s.logAudit(c, "delete", id, p, nil)
	return nil
}

func (s *ProjectService) logAudit(c *gin.Context, action, id string, old, new interface{}) {
	utils.LogAuditWithConsole(c, action, "project", fmt.Sprintf("p_id=%s", id), old, new, "", s.Repos.Audit)
}

func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func getValueInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
