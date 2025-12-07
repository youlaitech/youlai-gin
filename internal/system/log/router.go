package log

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/system/log/handler"
)

// RegisterRoutes 注册日志管理路由
func RegisterRoutes(router *gin.RouterGroup) {
	handler.RegisterRoutes(router)
}
