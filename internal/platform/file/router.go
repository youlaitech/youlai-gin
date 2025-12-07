package file

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/platform/file/handler"
)

// RegisterRoutes 注册文件管理路由
func RegisterRoutes(router *gin.RouterGroup) {
	fileGroup := router.Group("/files")
	{
		fileGroup.POST("", handler.UploadFile)        // 单文件上传
		fileGroup.POST("/batch", handler.UploadFiles) // 批量上传
		fileGroup.POST("/image", handler.UploadImage) // 图片上传
		fileGroup.DELETE("", handler.DeleteFile)      // 删除文件
	}
}
