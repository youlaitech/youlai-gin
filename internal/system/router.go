package system

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/common/database"

	configHandler "youlai-gin/internal/system/config/handler"
	configRepo "youlai-gin/internal/system/config/repository"
	configService "youlai-gin/internal/system/config/service"
	deptHandler "youlai-gin/internal/system/dept/handler"
	deptRepo "youlai-gin/internal/system/dept/repository"
	deptService "youlai-gin/internal/system/dept/service"
	dictHandler "youlai-gin/internal/system/dict/handler"
	dictRepo "youlai-gin/internal/system/dict/repository"
	dictService "youlai-gin/internal/system/dict/service"
	logHandler "youlai-gin/internal/system/log/handler"
	logRepo "youlai-gin/internal/system/log/repository"
	logService "youlai-gin/internal/system/log/service"
	menuHandler "youlai-gin/internal/system/menu/handler"
	menuRepo "youlai-gin/internal/system/menu/repository"
	menuService "youlai-gin/internal/system/menu/service"
	noticeHandler "youlai-gin/internal/system/notice/handler"
	noticeRepo "youlai-gin/internal/system/notice/repository"
	noticeService "youlai-gin/internal/system/notice/service"
	roleHandler "youlai-gin/internal/system/role/handler"
	roleRepo "youlai-gin/internal/system/role/repository"
	roleService "youlai-gin/internal/system/role/service"
	userHandler "youlai-gin/internal/system/user/handler"
	userRepo "youlai-gin/internal/system/user/repository"
	userService "youlai-gin/internal/system/user/service"
)

// RegisterRoutes 装配系统管理各模块（部门/字典/菜单/配置/通知/日志/角色/用户）的路由。
// 所有模块均走 repository → service → handler 的依赖注入。
func RegisterRoutes(r *gin.RouterGroup) {
	db := database.DB

	roleSvc := roleService.NewService(roleRepo.NewRepository(db), userRepo.NewRepository(db))
	userSvc := userService.NewService(userRepo.NewRepository(db), roleRepo.NewRepository(db), roleSvc, deptRepo.NewRepository(db))

	deptHandler.NewHandler(deptService.NewService(deptRepo.NewRepository(db))).RegisterRoutes(r)
	dictHandler.NewHandler(dictService.NewService(dictRepo.NewRepository(db))).RegisterRoutes(r)
	menuHandler.NewHandler(menuService.NewService(menuRepo.NewRepository(db), roleSvc)).RegisterRoutes(r)
	configHandler.NewHandler(configService.NewService(configRepo.NewRepository(db))).RegisterRoutes(r)
	noticeHandler.NewHandler(noticeService.NewService(noticeRepo.NewRepository(db))).RegisterRoutes(r)
	logHandler.NewHandler(logService.NewService(logRepo.NewRepository(db))).RegisterRoutes(r)
	roleHandler.NewHandler(roleSvc).RegisterRoutes(r)
	userHandler.NewHandler(userSvc).RegisterRoutes(r)
}
