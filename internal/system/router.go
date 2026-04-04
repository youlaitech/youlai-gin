package system

import (
	"github.com/gin-gonic/gin"

	configHandler "youlai-gin/internal/system/config/handler"
	deptHandler "youlai-gin/internal/system/dept/handler"
	dictHandler "youlai-gin/internal/system/dict/handler"
	logHandler "youlai-gin/internal/system/log/handler"
	menuHandler "youlai-gin/internal/system/menu/handler"
	noticeHandler "youlai-gin/internal/system/notice/handler"
	roleHandler "youlai-gin/internal/system/role/handler"
	userHandler "youlai-gin/internal/system/user/handler"
)

// RegisterRoutes 注册系统管理模块所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	userHandler.RegisterUserRoutes(r)    // 用户管理
	roleHandler.RegisterRoleRoutes(r)     // 角色管理
	menuHandler.RegisterMenuRoutes(r)     // 菜单管理
	deptHandler.RegisterDeptRoutes(r)     // 部门管理
	dictHandler.RegisterDictRoutes(r)    // 字典管理
	configHandler.RegisterRoutes(r)       // 配置管理
	noticeHandler.RegisterRoutes(r)       // 通知公告
	logHandler.RegisterRoutes(r)         // 日志管理
}
