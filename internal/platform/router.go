package platform

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/ai"
	"youlai-gin/internal/platform/codegen"
	"youlai-gin/internal/platform/file"
)

// RegisterRoutes 注册平台服务模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	ai.RegisterRoutes(r)      // AI 助手
	codegen.RegisterRoutes(r) // 代码生成
	file.RegisterRoutes(r) // 文件管理
}
