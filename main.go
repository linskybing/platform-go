package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/config"
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/middleware"
	"github.com/linskybing/platform-go/routes"
)

func main() {
	config.LoadConfig()
	db.Init()
	middleware.Init()

	r := gin.Default()
	routes.RegisterRoutes(r)
	addr := fmt.Sprintf(":%s", config.ServerPort)
	r.Run(addr)
}
