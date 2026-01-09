package codegen

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/codegen/handler"
)

// RegisterRoutes 注册代码生成路由
func RegisterRoutes(router *gin.RouterGroup) {
	codegenGroup := router.Group("/codegen")
	{
		codegenGroup.GET("/table", handler.GetTablePage)
		codegenGroup.GET("/:tableName/config", handler.GetGenConfig)
		codegenGroup.POST("/:tableName/config", handler.SaveGenConfig)
		codegenGroup.DELETE("/:tableName/config", handler.DeleteGenConfig)
		codegenGroup.GET("/:tableName/preview", handler.GetPreview)
		codegenGroup.GET("/:tableName/download", handler.Download)
	}
}
