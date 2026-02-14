package project

import (
	"context"
	"errors"
	"time"

	domProject "github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/view"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"gorm.io/gorm"
)

const projectCacheTTL = 5 * time.Minute

// ProjectService provides basic project management operations.
type ProjectService struct {
	Repos *repository.Repos
	cache *cache.Service
}

// NewProjectService creates a new project service instance.
func NewProjectService(repos *repository.Repos, cacheSvc *cache.Service) *ProjectService {
	return &ProjectService{Repos: repos, cache: cacheSvc}
}

// GetProject retrieves a single project node by its unique identifier.
func (s *ProjectService) GetProject(id string) (*domProject.Project, error) {
	ctx := context.Background()
	if s.cache != nil && s.cache.Enabled() {
		var n domProject.Project
		if err := s.cache.GetJSON(ctx, projectByIDKey(id), &n); err == nil {
			return &n, nil
		}
	}
	n, err := s.Repos.Project.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(ctx, projectByIDKey(id), n, projectCacheTTL)
	}
	return n, nil
}

// GetProjectsByUser lists all projects (nodes) accessible to a specific user via group membership.
func (s *ProjectService) GetProjectsByUser(userID string) ([]view.ProjectUserView, error) {
	ctx := context.Background()
	// 1. Get User's Groups
	groups, err := s.Repos.Group.ListGroupsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return []view.ProjectUserView{}, nil
	}

	// 2. Extract Group Owner IDs (ResourceOwner IDs)
	var ownerIDs []string
	groupMap := make(map[string]string)
	for _, g := range groups {
		ownerIDs = append(ownerIDs, g.ID)
		groupMap[g.ID] = g.Name
	}

	// 3. Find descendant projects
	nodes, err := s.Repos.Project.ListDescendantProjects(ctx, ownerIDs)
	if err != nil {
		return nil, err
	}

	// 4. Map to View
	var views []view.ProjectUserView
	for _, n := range nodes {
		v := view.ProjectUserView{
			PID:         n.ID,
			ProjectName: n.Name,
		}
		if n.ParentID != nil {
			v.GID = *n.ParentID
			if name, ok := groupMap[*n.ParentID]; ok {
				v.GroupName = name
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// ListProjects returns all project nodes in the system (Admin only typically).
func (s *ProjectService) ListProjects() ([]domProject.Project, error) {
	return s.Repos.Project.ListProjects(context.Background())
}

// GroupProjectsByGID organizes project views by group id for API responses.
func (s *ProjectService) GroupProjectsByGID(records []view.ProjectUserView) map[string]map[string]interface{} {
	if len(records) == 0 {
		return nil
	}
	grouped := make(map[string]map[string]interface{})
	for _, rec := range records {
		group, ok := grouped[rec.GID]
		if !ok {
			group = map[string]interface{}{
				"group_id":   rec.GID,
				"group_name": rec.GroupName,
				"projects":   []map[string]string{},
			}
		}
		projects := group["projects"].([]map[string]string)
		projects = append(projects, map[string]string{
			"p_id":         rec.PID,
			"project_name": rec.ProjectName,
		})
		group["projects"] = projects
		grouped[rec.GID] = group
	}
	return grouped
}
