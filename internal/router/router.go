package router

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth"
	"youlai-gin/internal/user"
	pkgAuth "youlai-gin/pkg/auth"
)

// Register 注册所有业务路由
func Register(r *gin.Engine, tokenManager pkgAuth.TokenManager) {
	api := r.Group("/api/v1")

	// 认证模块
	auth.RegisterRoutes(api, tokenManager)

	// 用户模块
	user.RegisterRoutes(api)
}
