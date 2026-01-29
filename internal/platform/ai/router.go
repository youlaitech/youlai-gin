package ai

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/ai/handler"
)

// RegisterRoutes 注册 AI 助手路由
func RegisterRoutes(router *gin.RouterGroup) {
	aiGroup := router.Group("/ai/assistant")
	{
		aiGroup.POST("/parse", handler.ParseCommand)
		aiGroup.POST("/execute", handler.ExecuteCommand)
		aiGroup.GET("/records", handler.GetRecordPage)
	}
}
