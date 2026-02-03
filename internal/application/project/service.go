package project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/view"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/utils"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectService struct {
	Repos *repository.Repos
	cache *cache.Service
}

func NewProjectService(repos *repository.Repos) *ProjectService {
	return NewProjectServiceWithCache(repos, nil)
}

func NewProjectServiceWithCache(repos *repository.Repos, cacheSvc *cache.Service) *ProjectService {
	return &ProjectService{
		Repos: repos,
		cache: cacheSvc,
	}
}

const projectCacheTTL = 5 * time.Minute

func (s *ProjectService) GetProject(id uint) (*project.Project, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached project.Project
		if err := s.cache.GetJSON(context.Background(), projectByIDKey(id), &cached); err == nil {
			return &cached, nil
		}
	}

	p, err := s.Repos.Project.GetProjectByID(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), projectByIDKey(id), p, projectCacheTTL)
	}
	return &p, nil
}

func (s *ProjectService) GetProjectsByUser(id uint) ([]view.ProjectUserView, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []view.ProjectUserView
		if err := s.cache.GetJSON(context.Background(), projectByUserKey(id), &cached); err == nil {
			return cached, nil
		}
	}

	p, err := s.Repos.Project.ListProjectsByUserID(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), projectByUserKey(id), p, projectCacheTTL)
	}
	return p, nil
}

func (s *ProjectService) GroupProjectsByGID(records []view.ProjectUserView) map[string]map[string]interface{} {
	grouped := make(map[string]map[string]interface{})

	for _, r := range records {
		key := strconv.Itoa(int(r.GID))
		if _, exists := grouped[key]; !exists {
			grouped[key] = map[string]interface{}{
				"GroupName": r.GroupName,
				"Projects":  []map[string]interface{}{},
			}
		}
		projects := grouped[key]["Projects"].([]map[string]interface{})
		projects = append(projects, map[string]interface{}{
			"PID":         r.PID,
			"ProjectName": r.ProjectName,
		})
		grouped[key]["Projects"] = projects
	}

	return grouped
}

func (s *ProjectService) CreateProject(c *gin.Context, input project.CreateProjectDTO) (*project.Project, error) {
	// Validate that the group exists
	if _, err := s.Repos.Group.GetGroupByID(input.GID); err != nil {
		return nil, fmt.Errorf("group with ID %d not found", input.GID)
	}

	p := &project.Project{
		ProjectName: input.ProjectName,
		GID:         input.GID,
	}
	if input.Description != nil {
		p.Description = *input.Description
	}
	if input.GPUQuota != nil {
		p.GPUQuota = *input.GPUQuota
	}
	err := s.Repos.Project.CreateProject(p)
	if err != nil {
		return nil, err
	}
	s.invalidateProjectCache(p.PID)

	// Sanity check: Verify GORM properly populated the PID
	if p.PID == 0 {
		fmt.Fprintf(os.Stderr, "ERROR CreateProject: GORM did not populate p.PID after CREATE. This indicates a database or driver issue.\n")
		return nil, fmt.Errorf("failed to get project ID from database")
	}

	logFn := utils.LogAuditWithConsole
	go func(fn func(*gin.Context, string, string, string, interface{}, interface{}, string, repository.AuditRepo)) {
		fn(c, "create", "project", fmt.Sprintf("p_id=%d", p.PID), nil, p, "", s.Repos.Audit)
	}(logFn)

	return p, nil
}

func (s *ProjectService) UpdateProject(c *gin.Context, id uint, input project.UpdateProjectDTO) (*project.Project, error) {
	p, err := s.Repos.Project.GetProjectByID(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	oldProject := p

	if input.ProjectName != nil {
		p.ProjectName = *input.ProjectName
	}
	if input.Description != nil {
		p.Description = *input.Description
	}
	if input.GID != nil {
		p.GID = *input.GID
	}
	if input.GPUQuota != nil {
		p.GPUQuota = *input.GPUQuota
	}

	err = s.Repos.Project.UpdateProject(&p)
	if err == nil {
		s.invalidateProjectCache(p.PID)
		utils.LogAuditWithConsole(c, "update", "project", fmt.Sprintf("p_id=%d", p.PID), oldProject, p, "", s.Repos.Audit)
	}

	return &p, err
}

func (s *ProjectService) DeleteProject(c *gin.Context, id uint) error {
	project, err := s.Repos.Project.GetProjectByID(id)
	if err != nil {
		return ErrProjectNotFound
	}

	_ = s.RemoveProjectResources(id)

	err = s.Repos.Project.DeleteProject(id)
	if err == nil {
		s.invalidateProjectCache(project.PID)
		utils.LogAuditWithConsole(c, "delete", "project", fmt.Sprintf("p_id=%d", project.PID), project, nil, "", s.Repos.Audit)
	}
	return err
}

func (s *ProjectService) ListProjects() ([]project.Project, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []project.Project
		if err := s.cache.GetJSON(context.Background(), projectListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	projects, err := s.Repos.Project.ListProjects()
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), projectListKey(), projects, projectCacheTTL)
	}
	return projects, nil
}

func (s *ProjectService) RemoveProjectResources(projectID uint) error {
	project, err := s.Repos.Project.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project info: %w", err)
	}

	users, err := s.Repos.User.ListUsersByProjectID(projectID)
	if err != nil {
		return err
	}

	for _, user := range users {
		safeUsername := k8s.ToSafeK8sName(user.Username)
		ns := k8s.FormatNamespaceName(projectID, safeUsername)

		if err := k8s.DeleteNamespace(ns); err != nil {
			slog.Error("failed to delete user instance namespace",
				"project_id", projectID,
				"username", user.Username,
				"namespace", ns,
				"error", err)
		}
	}

	groupStorageNs := k8s.GenerateSafeResourceName("group", project.ProjectName, project.PID)

	slog.Info("cleaning up group storage namespace",
		"project_id", projectID,
		"namespace", groupStorageNs)
	if err := k8s.DeleteNamespace(groupStorageNs); err != nil {
		return fmt.Errorf("failed to delete group storage namespace %s: %w", groupStorageNs, err)
	}

	return nil
}

func projectListKey() string {
	return "cache:project:list"
}

func projectByIDKey(id uint) string {
	return fmt.Sprintf("cache:project:by-id:%d", id)
}

func projectByUserKey(userID uint) string {
	return fmt.Sprintf("cache:project:by-user:%d", userID)
}

func (s *ProjectService) invalidateProjectCache(projectID uint) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, projectListKey(), projectByIDKey(projectID))
	_ = s.cache.InvalidatePrefix(ctx, "cache:project:by-user:")
}
