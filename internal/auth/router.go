package auth

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/handler"
	"youlai-gin/internal/auth/service"
	pkgAuth "youlai-gin/pkg/auth"
)

// RegisterRoutes 供上层 router 或 main 调用
func RegisterRoutes(api *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	// 初始化 TokenManager
	service.InitTokenManager(tokenManager)

	// 注册路由
	handler.RegisterAuthRoutes(api)
}
