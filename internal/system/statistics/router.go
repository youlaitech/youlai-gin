package statistics

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/statistics/handler"
)

// RegisterRoutes 注册统计分析路由
func RegisterRoutes(router *gin.RouterGroup) {
	handler.RegisterRoutes(router)
}







