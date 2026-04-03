package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	youlaDocs "youlai-gin/api"
	"youlai-gin/internal/router"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/config"
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/logger"
	"youlai-gin/pkg/middleware"
	"youlai-gin/pkg/redis"
	"youlai-gin/pkg/requestid"
	"youlai-gin/pkg/sse"

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
	youlaDocs.SwaggerInfo.Version = "4.1.0"
	r := gin.New()
	r.Use(requestid.Middleware())
	r.Use(logger.Middleware())
	r.Use(logger.Recovery())
	r.Use(middleware.ErrorHandler())

	// 全局限流中间件（每秒 10 个请求，突发 20 个）
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

	logger.Log.Sugar().Infof("服务启动在 :8000 [环境: %s]", config.GetEnv())

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
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
		log.Fatalf("服务器关闭失败: %v", err)
	}
	logger.Log.Sugar().Info("服务器已关闭")
}
