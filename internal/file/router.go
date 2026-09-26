package file

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/file/handler"
	"youlai-gin/internal/file/service"
)

// RegisterRoutes 注册文件管理路由（service → handler 依赖注入）
func RegisterRoutes(router *gin.RouterGroup) {
	handler.NewFileHandler(service.NewFileService()).RegisterRoutes(router)
}
