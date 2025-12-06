package main

import (
	"log"

	"github.com/gin-gonic/gin"

	_ "youlai-gin/docs"
	"youlai-gin/internal/database"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/router"
	"youlai-gin/pkg/logger"
	"youlai-gin/pkg/requestid"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 根据 Gin 模式选择日志配置文件
	configFile := "config/logger.yaml"
	if gin.Mode() == gin.ReleaseMode {
		configFile = "config/logger-prod.yaml"
	}
	
	if err := logger.InitFromYAML(configFile); err != nil {
		panic(err)
	}
	defer logger.Sync()

	database.Init()

	r := gin.New()
	r.Use(requestid.Middleware())
	r.Use(logger.Middleware())
	r.Use(logger.Recovery())
	r.Use(middleware.ErrorHandler())

	// 业务路由统一注册
	router.Register(r)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
