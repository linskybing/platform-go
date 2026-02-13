package routes

import (
	"github.com/gin-gonic/gin"
	wsHandlers "github.com/linskybing/platform-go/internal/api/handlers/websocket"
	"github.com/linskybing/platform-go/internal/application"
)

func registerWebSocketRoutes(r *gin.RouterGroup, services *application.Services) {
	ws := r.Group("/ws")
	{
		ws.GET("/exec", wsHandlers.ExecWebSocketHandler)
		ws.GET("/watch/:namespace", wsHandlers.WatchNamespaceHandler)
		ws.GET("/pod-logs", wsHandlers.StreamPodLogsHandler)
		ws.GET("/job-status/:id", func(c *gin.Context) {
			wsHandlers.WatchJobStatusHandler(c, services.Repos)
		})
		// Job-related websocket endpoints are temporarily disabled.
		// ws.GET("/image-pull/:job_id", func(c *gin.Context) {
		//	wsHandlers.WatchImagePullHandler(c, services.Image)
		// })
		// ws.GET("/image-pull-all", func(c *gin.Context) {
		//	wsHandlers.WatchMultiplePullJobsHandler(c, services.Image)
		// })
	}
}
