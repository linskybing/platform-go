// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/src/config"
	"github.com/linskybing/platform-go/src/db"
	_ "github.com/linskybing/platform-go/src/docs"
	"github.com/linskybing/platform-go/src/k8sclient"
	"github.com/linskybing/platform-go/src/middleware"
	"github.com/linskybing/platform-go/src/routes"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.LoadConfig()
	config.InitK8sConfig()
	db.Init()
	// minio.InitMinio()
	k8sclient.Init()
	middleware.Init()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.SetTrustedProxies([]string{"127.0.0.1"})
	routes.RegisterRoutes(r)
	addr := fmt.Sprintf(":%s", config.ServerPort)
	r.Run(addr)
}
