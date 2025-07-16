package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/handlers"
	"github.com/linskybing/platform-go/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		projects := auth.Group("/projects")
		{
			projects.GET("", handlers.GetProjects)
			projects.GET("/:id", handlers.GetProjectByID)
			projects.POST("", handlers.CreateProject)
			projects.PUT("/:id", handlers.UpdateProject)
			projects.DELETE("/:id", handlers.DeleteProject)
		}
	}
}
