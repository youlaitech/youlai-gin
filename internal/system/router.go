package system

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/dept"
	"youlai-gin/internal/system/dict"
	"youlai-gin/internal/system/menu"
	"youlai-gin/internal/system/permission"
	"youlai-gin/internal/system/role"
	"youlai-gin/internal/system/user"
)

// RegisterRoutes 注册系统管理模块所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	user.RegisterRoutes(r)       // 用户管理
	role.RegisterRoutes(r)       // 角色管理
	menu.RegisterRoutes(r)       // 菜单管理
	dept.RegisterRoutes(r)       // 部门管理
	dict.RegisterRoutes(r)       // 字典管理
	permission.RegisterRoutes(r) // 权限管理
}
