package system

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/common/database"

	configHandler "youlai-gin/internal/system/config/handler"
	deptHandler "youlai-gin/internal/system/dept/handler"
	deptRepo "youlai-gin/internal/system/dept/repository"
	deptService "youlai-gin/internal/system/dept/service"
	dictHandler "youlai-gin/internal/system/dict/handler"
	logHandler "youlai-gin/internal/system/log/handler"
	menuHandler "youlai-gin/internal/system/menu/handler"
	noticeHandler "youlai-gin/internal/system/notice/handler"
	roleHandler "youlai-gin/internal/system/role/handler"
	userHandler "youlai-gin/internal/system/user/handler"
)

// RegisterRoutes 注册系统管理模块所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	userHandler.RegisterUserRoutes(r)
	roleHandler.RegisterRoleRoutes(r)
	menuHandler.RegisterMenuRoutes(r)
	initDeptHandler().RegisterRoutes(r)
	dictHandler.RegisterDictRoutes(r)
	configHandler.RegisterRoutes(r)
	noticeHandler.RegisterRoutes(r)
	logHandler.RegisterRoutes(r)
}

// initDeptHandler 手动组装部门模块依赖（后续迁移到 wire）
func initDeptHandler() *deptHandler.Handler {
	repo := deptRepo.NewRepository(database.DB)
	svc := deptService.NewService(repo)
	return deptHandler.NewHandler(svc)
}
