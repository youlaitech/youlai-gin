package role

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/role/handler"
)

// RegisterRoutes 注册角色模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	handler.RegisterRoleRoutes(r)
}
