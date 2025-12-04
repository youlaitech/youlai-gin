package user

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/user/handler"
	"youlai-gin/internal/user/service"
)

// RegisterRoutes 供上层 router 或 main 调用
func RegisterRoutes(api *gin.RouterGroup) {
	// 启动时自动迁移 users 表（也可以放到统一的 migrate 入口）
	_ = service.AutoMigrate()

	handler.RegisterUserRoutes(api)
}
