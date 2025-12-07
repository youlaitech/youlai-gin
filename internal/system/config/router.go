package config

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/system/config/handler"
)

// RegisterRoutes 注册配置管理路由
func RegisterRoutes(router *gin.RouterGroup) {
	handler.RegisterRoutes(router)
}
