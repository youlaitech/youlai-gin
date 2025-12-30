package main

import (
	"log"

	"github.com/gin-gonic/gin"

	_ "youlai-gin/docs"
	"youlai-gin/pkg/database"
	"youlai-gin/internal/health"
	"youlai-gin/internal/router"
	roleService "youlai-gin/internal/system/role/service"
	"youlai-gin/pkg/middleware"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/config"
	"youlai-gin/pkg/logger"
	pkgMiddleware "youlai-gin/pkg/middleware"
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

	// 5. 初始化角色权限缓存
	if err := roleService.InitRolePermsCache(); err != nil {
		log.Printf("警告: 角色权限缓存初始化失败: %v", err)
		// 不阻断启动，继续运行
	}

	// 6. 初始化 TokenManager
	tokenManager, err := auth.CreateTokenManager(&config.Cfg.Security)
	if err != nil {
		log.Fatalf("TokenManager 初始化失败: %v", err)
	}

	// 7. 启动 Gin 服务
	r := gin.New()
	r.Use(requestid.Middleware())
	r.Use(logger.Middleware())
	r.Use(logger.Recovery())
	r.Use(middleware.ErrorHandler())
	
	// 全局限流中间件（每秒 10 个请求，突发 20 个）
	r.Use(pkgMiddleware.RateLimitByIP())

	// 健康检查路由（无需认证）
	health.RegisterRoutes(r)

	// 业务路由统一注册
	router.Register(r, tokenManager)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Log.Sugar().Infof("服务启动在 :8000 [环境: %s]", config.GetEnv())
	if err := r.Run(":8000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
