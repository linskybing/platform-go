package application

import (
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/group"
	"github.com/linskybing/platform-go/internal/application/project"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

// Constructor wrappers for exported service types
// These allow tests and other packages to construct services directly from the application package

func NewConfigFileService(repos *repository.Repos) *configfile.ConfigFileService {
	return configfile.NewConfigFileService(repos)
}

func NewGroupService(repos *repository.Repos) *group.GroupService {
	return group.NewGroupService(repos)
}

func NewProjectService(repos *repository.Repos) *project.ProjectService {
	return project.NewProjectService(repos, (*cache.Service)(nil))
}

func NewUserGroupService(repos *repository.Repos) *group.UserGroupService {
	return group.NewUserGroupService(repos)
}
