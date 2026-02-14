package job

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers/gpuusage"
	"github.com/linskybing/platform-go/internal/api/handlers/job"
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/executor"
	appgpu "github.com/linskybing/platform-go/internal/application/gpuusage"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/events"
	"github.com/linskybing/platform-go/internal/plugin"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/prometheus"
)

type JobPlugin struct {
	handler    *job.JobHandler
	gpuHandler *gpuusage.GPUUsageHandler
	repos      *repository.Repos
	log        *slog.Logger
}

func NewJobPlugin() *JobPlugin {
	return &JobPlugin{}
}

func (p *JobPlugin) Name() string {
	return "job-manager"
}

func (p *JobPlugin) Version() string {
	return "1.0.0"
}

func (p *JobPlugin) Init(ctx plugin.PluginContext) error {
	p.log = ctx.Logger
	if p.log == nil {
		p.log = slog.Default()
	}

	repos := repository.NewRepositories(ctx.DB)
	p.repos = repos

	// Default to LocalExecutor
	exec := executor.NewLocalExecutor(repos)

	configSvc := configfile.NewConfigFileServiceWithExecutor(repos, ctx.Cache, exec)

	// Prometheus client
	promClient, err := prometheus.NewClient(config.PrometheusAddr)
	if err != nil {
		p.log.Warn("failed to initialize prometheus client for job plugin", "error", err)
	}

	p.handler = job.NewJobHandler(repos, exec, configSvc)

	gpuService := appgpu.NewGPUUsageService(repos, promClient)
	p.gpuHandler = gpuusage.NewGPUUsageHandler(gpuService)

	return nil
}

func (p *JobPlugin) RegisterRoutes(r *gin.RouterGroup) {
	jobs := r.Group("/jobs")
	{
		jobs.GET("/templates", p.handler.ListTemplates)
		jobs.POST("/submit", p.handler.SubmitJob)
		jobs.GET("", p.handler.ListJobs)

		jobs.GET("/:id", p.handler.GetJob)
		jobs.POST("/:id/cancel", p.handler.CancelJob)

		jobs.GET("/:id/gpu-usage", p.gpuHandler.GetJobGPUUsage)
		jobs.GET("/:id/gpu-summary", p.gpuHandler.GetJobGPUSummary)
	}
}

func (p *JobPlugin) RegisterMigrations() []plugin.Migration {
	return nil
}

func (p *JobPlugin) RegisterEvents(bus events.EventBus) {
}

func (p *JobPlugin) Shutdown() error {
	return nil
}
