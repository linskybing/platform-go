package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers/audit"
	"github.com/linskybing/platform-go/internal/api/handlers/cluster"
	"github.com/linskybing/platform-go/internal/api/handlers/configfile"
	"github.com/linskybing/platform-go/internal/api/handlers/filebrowser"
	"github.com/linskybing/platform-go/internal/api/handlers/form"
	"github.com/linskybing/platform-go/internal/api/handlers/gpuusage"
	"github.com/linskybing/platform-go/internal/api/handlers/group"
	"github.com/linskybing/platform-go/internal/api/handlers/groupstorage"
	"github.com/linskybing/platform-go/internal/api/handlers/image"
	"github.com/linskybing/platform-go/internal/api/handlers/job"
	"github.com/linskybing/platform-go/internal/api/handlers/k8s"
	"github.com/linskybing/platform-go/internal/api/handlers/notification"
	"github.com/linskybing/platform-go/internal/api/handlers/project"
	"github.com/linskybing/platform-go/internal/api/handlers/pvcbinding"
	"github.com/linskybing/platform-go/internal/api/handlers/storageperm"
	"github.com/linskybing/platform-go/internal/api/handlers/user"
	"github.com/linskybing/platform-go/internal/api/handlers/userstorage"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

type Handlers struct {
	Audit        *audit.AuditHandler
	ConfigFile   *configfile.ConfigFileHandler
	Group        *group.GroupHandler
	Form         *form.FormHandler
	Image        *image.ImageHandler
	Job          *job.JobHandler
	GPUUsage     *gpuusage.GPUUsageHandler
	Cluster      *cluster.ClusterHandler
	PVCBinding   *pvcbinding.PVCBindingHandler
	FileBrowser  *filebrowser.FileBrowserHandler
	StoragePerm  *storageperm.StoragePermissionHandler
	GroupStorage *groupstorage.GroupStorageHandler
	UserStorage  *userstorage.UserStorageHandler
	Project      *project.ProjectHandler
	User         *user.UserHandler
	UserGroup    *user.UserGroupHandler
	K8s          *k8s.K8sHandler
	Notification *notification.NotificationHandler
	Router       *gin.Engine
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
		Audit:        audit.NewAuditHandler(svc.Audit),
		ConfigFile:   configfile.NewConfigFileHandler(svc.ConfigFile),
		Group:        group.NewGroupHandler(svc.Group),
		Form:         form.NewFormHandler(svc.Form),
		Image:        image.NewImageHandlerWithCache(svc.Image, cacheSvc, logger),
		Job:          job.NewJobHandler(repos, svc.ConfigFile.GetExecutor(), svc.ConfigFile),
		GPUUsage:     gpuusage.NewGPUUsageHandler(svc.GPUUsage),
		Cluster:      cluster.NewClusterHandler(svc.Cluster),
		PVCBinding:   pvcbinding.NewPVCBindingHandler(svc.K8s.PVCBindingManager),
		FileBrowser:  filebrowser.NewFileBrowserHandler(svc.K8s.FileBrowserManager),
		StoragePerm:  storageperm.NewStoragePermissionHandler(svc.K8s.PermissionManager),
		GroupStorage: groupstorage.NewGroupStorageHandler(svc.K8s.StorageManager, svc.K8s.FileBrowserManager, svc.K8s.PermissionManager, repos.Audit),
		UserStorage:  userstorage.NewUserStorageHandler(svc.K8s, repos.Audit),
		Project:      project.NewProjectHandler(svc.Project),
		User:         user.NewUserHandler(svc.User),
		UserGroup:    user.NewUserGroupHandler(svc.UserGroup),
		K8s:          k8s.NewK8sHandler(svc.K8s, svc.User, svc.Project),
		Notification: notification.NewNotificationHandler(),
		Router:       router,
	}
	return h
}
