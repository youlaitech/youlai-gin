package main

import (
	"log"

	"github.com/gin-gonic/gin"

	_ "youlai-gin/docs"
	"youlai-gin/internal/database"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/router"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/config"
	"youlai-gin/pkg/logger"
	"youlai-gin/pkg/redis"
	"youlai-gin/pkg/requestid"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 1. 加载配置（根据 APP_ENV 环境变量或默认 dev）
	if err := config.Load(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 初始化日志
	logger.InitWithConfig(&config.Cfg.Logger)
	defer logger.Sync()

	// 3. 初始化数据库
	if err := database.InitWithConfig(&config.Cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 4. 初始化 Redis
	if err := redis.InitWithConfig(&config.Cfg.Redis); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}

	// 5. 初始化 TokenManager
	tokenManager, err := auth.CreateTokenManager(&config.Cfg.Security)
	if err != nil {
		log.Fatalf("TokenManager 初始化失败: %v", err)
	}

	// 6. 启动 Gin 服务
	r := gin.New()
	r.Use(requestid.Middleware())
	r.Use(logger.Middleware())
	r.Use(logger.Recovery())
	r.Use(middleware.ErrorHandler())

	// 业务路由统一注册
	router.Register(r, tokenManager)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Log.Sugar().Infof("服务启动在 :8000 [环境: %s]", config.GetEnv())
	if err := r.Run(":8000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
