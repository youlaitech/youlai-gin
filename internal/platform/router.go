package platform

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/platform/file"
)

// RegisterRoutes 注册平台服务模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	file.RegisterRoutes(r) // 文件管理
}
