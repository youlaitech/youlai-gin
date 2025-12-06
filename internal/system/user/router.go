package user

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/user/handler"
)

// RegisterRoutes 注册用户模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	handler.RegisterUserRoutes(r)
}
