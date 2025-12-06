package menu

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/menu/handler"
)

// RegisterRoutes 注册菜单模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	handler.RegisterMenuRoutes(r)
}
