package application

import (
	"log"

	"github.com/linskybing/platform-go/internal/application/audit"
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/group"
	"github.com/linskybing/platform-go/internal/application/image"
	appk8s "github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/application/project"
	"github.com/linskybing/platform-go/internal/application/user"
	"github.com/linskybing/platform-go/internal/repository"
)

type Services struct {
	Audit      *audit.AuditService
	ConfigFile *configfile.ConfigFileService
	Group      *group.GroupService
	Project    *project.ProjectService
	Resource   *ResourceService
	UserGroup  *group.UserGroupService
	User       *user.UserService
	K8s        *appk8s.K8sService
	Form       *FormService
	Image      *image.ImageService
}

func New(repos *repository.Repos) *Services {
	imageService := image.NewImageService(repos.Image)
	k8sService, err := appk8s.NewK8sService(repos, imageService)
	if err != nil {
		log.Fatalf("failed to initialize K8sService: %v", err)
	}

	return &Services{
		Audit:      audit.NewAuditService(repos),
		ConfigFile: configfile.NewConfigFileService(repos),
		Group:      group.NewGroupService(repos),
		Project:    project.NewProjectService(repos),
		Resource:   NewResourceService(repos),
		UserGroup:  group.NewUserGroupService(repos),
		User:       user.NewUserService(repos),
		K8s:        k8sService,
		Form:       NewFormService(repos.Form),
		Image:      imageService,
	}
}
