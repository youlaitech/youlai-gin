package router

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth"
	"youlai-gin/internal/platform"
	"youlai-gin/internal/system"
	pkgAuth "youlai-gin/pkg/auth"
)

// Register 注册所有业务路由
func Register(r *gin.Engine, tokenManager pkgAuth.TokenManager) {
	api := r.Group("/api/v1")

	// 认证模块（无需认证）
	auth.RegisterRoutes(api, tokenManager)

	// 需要认证的路由组
	authorized := api.Group("")
	authorized.Use(pkgAuth.Middleware(tokenManager))
	{
		// 系统管理模块（包含用户、角色、菜单、部门、字典、配置、通知、日志）
		system.RegisterRoutes(authorized)

		// 平台服务模块（包含文件上传）
		platform.RegisterRoutes(authorized)
	}
}
