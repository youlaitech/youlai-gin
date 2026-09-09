package message

import (
	"github.com/gin-gonic/gin"

	pkgAuth "youlai-gin/internal/common/auth"
	"youlai-gin/internal/message/handler"
)

// RegisterRoutes 注册 SSE 连接路由（handler → service → model 依赖注入）
func RegisterRoutes(r *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	handler.NewSseHandler(tokenManager).RegisterRoutes(r)
}