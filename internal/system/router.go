package system

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/common/database"

	configHandler "youlai-gin/internal/system/config/handler"
	deptHandler "youlai-gin/internal/system/dept/handler"
	deptRepo "youlai-gin/internal/system/dept/repository"
	deptService "youlai-gin/internal/system/dept/service"
	dictHandler "youlai-gin/internal/system/dict/handler"
	dictRepo "youlai-gin/internal/system/dict/repository"
	dictService "youlai-gin/internal/system/dict/service"
	logHandler "youlai-gin/internal/system/log/handler"
	menuHandler "youlai-gin/internal/system/menu/handler"
	menuRepo "youlai-gin/internal/system/menu/repository"
	menuService "youlai-gin/internal/system/menu/service"
	noticeHandler "youlai-gin/internal/system/notice/handler"
	roleHandler "youlai-gin/internal/system/role/handler"
	userHandler "youlai-gin/internal/system/user/handler"
)

// RegisterRoutes 装配系统管理各模块（部门/字典/菜单/角色/用户/配置/通知/日志）的路由。
// 部门、字典、菜单走 repository → service → handler 的依赖注入；其余模块自行注册。
func RegisterRoutes(r *gin.RouterGroup) {
	db := database.DB

	deptHandler.NewHandler(deptService.NewService(deptRepo.NewRepository(db))).RegisterRoutes(r)
	dictHandler.NewHandler(dictService.NewService(dictRepo.NewRepository(db))).RegisterRoutes(r)
	menuHandler.NewHandler(menuService.NewService(menuRepo.NewRepository(db))).RegisterRoutes(r)

	roleHandler.RegisterRoleRoutes(r)
	userHandler.RegisterUserRoutes(r)
	configHandler.RegisterRoutes(r)
	noticeHandler.RegisterRoutes(r)
	logHandler.RegisterRoutes(r)
}
