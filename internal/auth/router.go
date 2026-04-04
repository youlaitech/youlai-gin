package auth

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/handler"
	"youlai-gin/internal/auth/service"
	pkgAuth "youlai-gin/internal/common/auth"
)

// RegisterRoutes 供上层 router 或 main 调用
func RegisterRoutes(api *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	// 初始化 TokenManager
	service.InitTokenManager(tokenManager)

	// 初始化微信配置
	service.InitWechatConfig()

	// 注册认证路由
	handler.RegisterAuthRoutes(api)

	// 注册微信小程序认证路由
	handler.RegisterWxMaRoutes(api)
}
