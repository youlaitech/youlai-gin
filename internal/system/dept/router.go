package dept

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/dept/handler"
)

// RegisterRoutes 注册部门模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	handler.RegisterDeptRoutes(r)
}
