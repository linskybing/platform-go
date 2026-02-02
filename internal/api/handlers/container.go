package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers/image"
	"github.com/linskybing/platform-go/internal/api/handlers/k8s"
	"github.com/linskybing/platform-go/internal/api/handlers/project"
	"github.com/linskybing/platform-go/internal/api/handlers/user"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
)

type Handlers struct {
	Audit           *AuditHandler
	ConfigFile      *ConfigFileHandler
	Group           *GroupHandler
	Project         *project.QueryHandler
	ProjectMutate   *project.MutateHandler
	Resource        *ResourceHandler
	UserGroup       *user.GroupHandler
	UserGroupMutate *user.GroupMutateHandler
	User            *user.CrudHandler
	UserAuth        *user.AuthHandler
	K8s             *k8s.K8sHandler
	Form            *FormHandler
	Image           *image.ManagementHandler
	ImageRequest    *image.RequestHandler
	ImageApproval   *image.ApprovalHandler
	ImagePull       *image.PullHandler
	Router          *gin.Engine
}

func New(svc *application.Services, repos *repository.Repos, router *gin.Engine) *Handlers {
	h := &Handlers{
		Audit:           NewAuditHandler(svc.Audit),
		ConfigFile:      NewConfigFileHandler(svc.ConfigFile),
		Group:           NewGroupHandler(svc.Group),
		Project:         project.NewQueryHandler(svc.Project),
		ProjectMutate:   project.NewMutateHandler(svc.Project),
		Resource:        NewResourceHandler(svc.Resource),
		UserGroup:       user.NewGroupHandler(svc.UserGroup),
		UserGroupMutate: user.NewGroupMutateHandler(svc.UserGroup),
		User:            user.NewCrudHandler(svc.User),
		UserAuth:        user.NewAuthHandler(svc.User),
		K8s:             k8s.NewK8sHandler(svc.K8s, svc.User, svc.Project),
		Form:            NewFormHandler(svc.Form),
		Image:           image.NewManagementHandler(svc.Image),
		ImageRequest:    image.NewRequestHandler(svc.Image),
		ImageApproval:   image.NewApprovalHandler(svc.Image),
		ImagePull:       image.NewPullHandler(svc.Image),
		Router:          router,
	}
	return h
}
