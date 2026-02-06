package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

// TODO: Refactor Handler structure - implement sub-packages: image, k8s, project, user
// These handlers are currently in the same package and need to be reorganized

type Handlers struct {
	Audit       *AuditHandler
	ConfigFile  *ConfigFileHandler
	Group       *GroupHandler
	Form        *FormHandler
	Image       *ImageHandler
	PVCBinding  *PVCBindingHandler
	FileBrowser *FileBrowserHandler
	StoragePerm *StoragePermissionHandler
	Project     *ProjectHandler
	User        *UserHandler
	UserGroup   *UserGroupHandler
	K8s         *K8sHandler
	Router      *gin.Engine
}

// New creates a new Handlers instance without cache support.
func New(svc *application.Services, repos *repository.Repos, router *gin.Engine) *Handlers {
	return NewWithCache(svc, repos, router, nil, nil)
}

// NewWithCache creates a new Handlers instance with optional cache and logger support.
func NewWithCache(svc *application.Services, repos *repository.Repos, router *gin.Engine, cacheSvc *cache.Service, logger *slog.Logger) *Handlers {
	if logger == nil {
		logger = slog.Default()
	}

	h := &Handlers{
		Audit:       NewAuditHandler(svc.Audit),
		ConfigFile:  NewConfigFileHandler(svc.ConfigFile),
		Group:       NewGroupHandler(svc.Group),
		Form:        NewFormHandler(svc.Form),
		Image:       NewImageHandlerWithCache(svc.Image, cacheSvc, logger),
		PVCBinding:  NewPVCBindingHandler(svc.K8s.PVCBindingManager),
		FileBrowser: NewFileBrowserHandler(svc.K8s.FileBrowserManager),
		StoragePerm: NewStoragePermissionHandler(svc.K8s.PermissionManager),
		Project:     NewProjectHandler(svc.Project),
		User:        NewUserHandler(svc.User),
		UserGroup:   NewUserGroupHandler(svc.UserGroup),
		K8s:         NewK8sHandler(svc.K8s, svc.User, svc.Project),
		Router:      router,
	}
	return h
}
