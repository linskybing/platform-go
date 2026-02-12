package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/handlers"
	"github.com/linskybing/platform-go/internal/api/middleware"
)

func registerK8sRoutes(r *gin.RouterGroup, h *handlers.Handlers, authMw *middleware.AuthMiddleware) {
	k8sGroup := r.Group("/k8s")
	{
		// Pod Logs
		k8sGroup.GET("/namespaces/:ns/pods/:name/logs", h.K8s.GetPodLogs)

		// User Storage (OpenMyDrive)
		userStorage := k8sGroup.Group("/user-storage")
		{
			userStorage.GET("/status", h.K8s.CheckMyStorageStatus)
			userStorage.POST("/browse", h.K8s.OpenMyDrive)
			userStorage.DELETE("/browse", h.K8s.StopMyDrive)
			// Proxy routes
			userStorage.Any("/proxy/*path", h.K8s.UserStorageProxy)
		}
	}
}
