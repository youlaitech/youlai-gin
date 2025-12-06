package dict

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/dict/handler"
)

// RegisterRoutes 注册字典模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	handler.RegisterDictRoutes(r)
}
