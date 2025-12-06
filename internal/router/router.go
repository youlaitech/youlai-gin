package router

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth"
	"youlai-gin/internal/system"
	pkgAuth "youlai-gin/pkg/auth"
)

// Register 注册所有业务路由
func Register(r *gin.Engine, tokenManager pkgAuth.TokenManager) {
	api := r.Group("/api/v1")

	// 认证模块
	auth.RegisterRoutes(api, tokenManager)

	// 系统管理模块（包含用户、角色、菜单、部门、字典）
	system.RegisterRoutes(api)
}
