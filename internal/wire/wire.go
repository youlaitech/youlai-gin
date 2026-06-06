//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/config"
	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/redis"
	deptHandler "youlai-gin/internal/system/dept/handler"
	deptRepo "youlai-gin/internal/system/dept/repository"
	deptService "youlai-gin/internal/system/dept/service"
	dictHandler "youlai-gin/internal/system/dict/handler"
	dictRepo "youlai-gin/internal/system/dict/repository"
	dictService "youlai-gin/internal/system/dict/service"
	menuHandler "youlai-gin/internal/system/menu/handler"
	menuRepo "youlai-gin/internal/system/menu/repository"
	menuService "youlai-gin/internal/system/menu/service"
	roleHandler "youlai-gin/internal/system/role/handler"
	roleRepo "youlai-gin/internal/system/role/repository"
	roleService "youlai-gin/internal/system/role/service"
	userHandler "youlai-gin/internal/system/user/handler"
	userRepo "youlai-gin/internal/system/user/repository"
	userService "youlai-gin/internal/system/user/service"
)

// InfrastructureSet 基础设施注入
var InfrastructureSet = wire.NewSet(
	config.Load,
	database.NewDB,
	redis.NewClient,
	auth.CreateTokenManager,
)

// DeptSet 部门模块注入
var DeptSet = wire.NewSet(
	deptRepo.NewRepository,
	wire.Bind(new(deptService.Repository), new(*deptRepo.Repository)),
	deptService.NewService,
	deptHandler.NewHandler,
)

// DictSet 字典模块注入
var DictSet = wire.NewSet(
	dictRepo.NewRepository,
	wire.Bind(new(dictService.Repository), new(*dictRepo.Repository)),
	dictService.NewService,
	dictHandler.NewHandler,
)

// MenuSet 菜单模块注入
var MenuSet = wire.NewSet(
	menuRepo.NewRepository,
	wire.Bind(new(menuService.Repository), new(*menuRepo.Repository)),
	menuService.NewService,
	menuHandler.NewHandler,
)

// RoleSet 角色模块注入
var RoleSet = wire.NewSet(
	roleRepo.NewRepository,
	wire.Bind(new(roleService.Repository), new(*roleRepo.Repository)),
	roleService.NewService,
	roleHandler.NewHandler,
)

// UserSet 用户模块注入
var UserSet = wire.NewSet(
	userRepo.NewRepository,
	userService.NewService,
	userHandler.NewHandler,
)

// App 应用上下文
type App struct {
	DB           *gorm.DB
	RedisClient  *redis.Client
	TokenManager auth.TokenManager
	Cfg          *config.Config
	DeptHandler  *deptHandler.Handler
	DictHandler  *dictHandler.Handler
	MenuHandler  *menuHandler.Handler
	RoleHandler  *roleHandler.Handler
	UserHandler  *userHandler.Handler
}
