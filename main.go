package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	youlaDocs "youlai-gin/api"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/config"
	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/logger"
	"youlai-gin/internal/common/redis"
	"youlai-gin/internal/common/storage"
	sse "youlai-gin/internal/message/service"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/router"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const swaggerIndexHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="/swagger/swagger-ui.css" />
    <link rel="icon" type="image/png" href="/swagger/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="/swagger/favicon-16x16.png" sizes="16x16" />
    <style>
      html {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after {
        box-sizing: inherit;
      }
      body {
        margin: 0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="/swagger/swagger-ui-bundle.js"></script>
    <script src="/swagger/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function () {
        const ui = SwaggerUIBundle({
          url: "/swagger/doc.json",
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
          plugins: [SwaggerUIBundle.plugins.DownloadUrl],
          layout: "StandaloneLayout",
          tagsSorter: "alpha",
          operationsSorter: "alpha",
        });
        window.ui = ui;
      };
    </script>
  </body>
</html>
`

const Version = "4.2.0"

func main() {
	// 加载配置（APP_ENV 或默认 dev）
	if err := config.Load(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 初始化日志
	logger.InitWithConfig(&config.Cfg.Logger)
	defer logger.Sync()

	// 初始化数据库
	if err := database.InitWithConfig(&config.Cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化 Redis
	if err := redis.InitWithConfig(&config.Cfg.Redis); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}
	logger.Log.Sugar().Infof("Redis 已连接: %s:%d (db=%d)", config.Cfg.Redis.Host, config.Cfg.Redis.Port, config.Cfg.Redis.Database)

	// 初始化文件存储（按 storage.type 选择 s3 / local 驱动）
	fileCfg := &config.Cfg.FileStorage
	storageCfg := &storage.Config{
		Type:             fileCfg.Type,
		Endpoint:         fileCfg.S3.Endpoint,
		AccessKey:        fileCfg.S3.AccessKey,
		SecretKey:        fileCfg.S3.SecretKey,
		Bucket:           fileCfg.S3.Bucket,
		Domain:           fileCfg.S3.Domain,
		PathNoBucketName: fileCfg.S3.PathNoBucketName,
		BasePath:         fileCfg.Local.Path,
	}
	if storageCfg.Type == "" {
		storageCfg.Type = storage.TypeLocal
	}
	if storageCfg.Type == storage.TypeLocal {
		storageCfg.Domain = fileCfg.Local.BaseURL
	}
	if err := storage.InitDefaultStorage(storageCfg); err != nil {
		log.Fatalf("文件存储初始化失败: %v", err)
	}
	logger.Log.Sugar().Infof("文件存储已初始化: type=%s", storageCfg.Type)

	// 初始化 SSE 服务
	sse.InitSseService()

	// 初始化 TokenManager
	tokenManager, err := auth.CreateTokenManager(&config.Cfg.Security)
	if err != nil {
		log.Fatalf("TokenManager 初始化失败: %v", err)
	}

	// 启动 Gin 服务
	youlaDocs.SwaggerInfo.Title = "youlai-gin"
	youlaDocs.SwaggerInfo.Description = "youlai 全家桶（Go/Gin）权限管理后台接口文档"
	youlaDocs.SwaggerInfo.Version = Version
	r := gin.New()
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.Middleware())
	r.Use(logger.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RateLimitByIP())

	// 业务路由
	router.Register(r, tokenManager)

	// Swagger 文档路由
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Param("any")
		if path == "" || path == "/" || path == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, swaggerIndexHTML)
			return
		}
		swaggerHandler(c)
	})

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	logger.Log.Sugar().Infof("服务启动在 %s [环境: %s, 版本: %s]", addr, config.GetEnv(), Version)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Sugar().Info("正在关闭服务器...")

	// 主动断开所有 SSE 连接
	sse.GetSseService().CloseAll()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}
	logger.Log.Sugar().Info("服务器已关闭")
}
