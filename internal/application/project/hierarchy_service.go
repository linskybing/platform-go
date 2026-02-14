package project

import (
	"context"
	"github.com/linskybing/platform-go/internal/domain/project"
)

// HierarchyService manages the ltree-based tree structure of projects.
type HierarchyService struct {
	projectSvc *ProjectService
}

// NewHierarchyService creates a new hierarchy management service.
func NewHierarchyService(ps *ProjectService) *HierarchyService {
	return &HierarchyService{projectSvc: ps}
}

// GetSubtree retrieves all descendants of a given node using ltree.
func (h *HierarchyService) GetSubtree(ctx context.Context, rootID string) ([]project.Project, error) {
	return h.projectSvc.Repos.Project.GetSubtree(ctx, rootID)
}

// MoveNode updates a node's parent and re-calculates the ltree path.
func (h *HierarchyService) MoveNode(ctx context.Context, nodeID, newParentID string) error {
	return h.projectSvc.Repos.Project.MoveNode(ctx, nodeID, newParentID)
}

// GetAncestors retrieves all parents of a node up to the root.
func (h *HierarchyService) GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error) {
	return h.projectSvc.Repos.Project.GetAncestors(ctx, nodeID)
}
