package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerClusterRoutes(r *gin.RouterGroup, h *handlers.Handlers, _ *middleware.AuthMiddleware) {
	cluster := r.Group("/api/cluster")
	{
		cluster.GET("/summary", h.Cluster.GetClusterSummary)
		cluster.GET("/nodes", h.Cluster.ListClusterNodes)
		cluster.GET("/nodes/:name", h.Cluster.GetClusterNode)
		cluster.GET("/gpu-usage", h.Cluster.ListPodGPUUsage)
	}
}
