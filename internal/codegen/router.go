package codegen

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/codegen/handler"
	"youlai-gin/internal/codegen/repository"
	"youlai-gin/internal/codegen/service"
	"youlai-gin/internal/common/database"
)

// RegisterRoutes 注册代码生成路由（repository → service → handler 依赖注入）
func RegisterRoutes(router *gin.RouterGroup) {
	repo := repository.NewRepository(database.DB)
	handler.NewCodegenHandler(service.NewCodegenService(repo)).RegisterRoutes(router)
}
