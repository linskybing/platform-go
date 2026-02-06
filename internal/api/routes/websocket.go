package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/application"
)

func registerWebSocketRoutes(r *gin.RouterGroup, services *application.Services) {
	ws := r.Group("/ws")
	{
		ws.GET("/exec", handlers.ExecWebSocketHandler)
		ws.GET("/watch/:namespace", handlers.WatchNamespaceHandler)
		ws.GET("/pod-logs", handlers.StreamPodLogsHandler)
		ws.GET("/image-pull/:job_id", func(c *gin.Context) {
			handlers.WatchImagePullHandler(c, services.Image)
		})
		ws.GET("/image-pull-all", func(c *gin.Context) {
			handlers.WatchMultiplePullJobsHandler(c, services.Image)
		})
	}
}
