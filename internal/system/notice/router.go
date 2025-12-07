package notice

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/system/notice/handler"
)

// RegisterRoutes 注册通知公告路由
func RegisterRoutes(router *gin.RouterGroup) {
	handler.RegisterRoutes(router)
}
