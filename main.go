package main

import (
	"log"

	"github.com/gin-gonic/gin"

	_ "youlai-gin/docs"
	"youlai-gin/internal/database"
	"youlai-gin/internal/router"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	database.Init()

	r := gin.Default()

	// 业务路由统一注册
	router.Register(r)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
