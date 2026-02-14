package application

import (
	"log"

	"github.com/linskybing/platform-go/internal/application/audit"
	"github.com/linskybing/platform-go/internal/application/cluster"
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/application/gpuusage"
	"github.com/linskybing/platform-go/internal/application/group"
	"github.com/linskybing/platform-go/internal/application/image"
	appk8s "github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/application/project"
	"github.com/linskybing/platform-go/internal/application/user"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/prometheus"
)

type Services struct {
	Repos      *repository.Repos // Exposed for handlers that need direct access
	Audit      *audit.AuditService
	ConfigFile *configfile.ConfigFileService
	Cluster    *cluster.ClusterService
	GPUUsage   *gpuusage.GPUUsageService
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
	return NewWithCache(repos, nil)
}

func NewWithCache(repos *repository.Repos, cacheSvc *cache.Service) *Services {
	imageService := image.NewImageService(repos.Image)
	k8sService, err := appk8s.NewK8sService(repos, imageService, cacheSvc)
	if err != nil {
		log.Fatalf("failed to initialize K8sService: %v", err)
	}

	promClient, err := prometheus.NewClient(config.PrometheusAddr)
	if err != nil {
		log.Printf("warning: failed to initialize prometheus client: %v", err)
	}
	clusterService := cluster.NewClusterService(cacheSvc, promClient)
	gpuUsageService := gpuusage.NewGPUUsageService(repos, promClient)

	// Create executor based on config
	var exec executor.Executor
	if config.FlashSchedEnabled || config.ExecutorMode == "scheduler" {
		flashJobClient := k8s.NewFlashJobClient(k8s.DynamicClient)
		exec = executor.NewSchedulerExecutor(repos, flashJobClient, config.SchedulerName)
	} else {
		exec = executor.NewLocalExecutor(repos)
	}

	return &Services{
		Repos:      repos,
		Audit:      audit.NewAuditService(repos),
		ConfigFile: configfile.NewConfigFileServiceWithExecutor(repos, cacheSvc, exec),
		Cluster:    clusterService,
		GPUUsage:   gpuUsageService,
		Group:      group.NewGroupServiceWithCache(repos, cacheSvc),
		Project:    project.NewProjectService(repos, cacheSvc),
		Resource:   NewResourceService(repos),
		UserGroup:  group.NewUserGroupService(repos),
		User:       user.NewUserServiceWithCache(repos, cacheSvc),
		K8s:        k8sService,
		Form:       NewFormService(repos.Form),
		Image:      imageService,
	}
}
