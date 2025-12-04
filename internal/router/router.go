package router

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/user"
)

// Register 注册所有业务路由
func Register(r *gin.Engine) {
	api := r.Group("/api/v1")

	// 用户模块
	user.RegisterRoutes(api)
}
