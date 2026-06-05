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
)

// InfrastructureSet 基础设施层 Provider Set
var InfrastructureSet = wire.NewSet(
	config.Load,
	database.NewDB,
	redis.NewClient,
	auth.CreateTokenManager,
)

// DeptSet 部门模块全套 Provider Set
var DeptSet = wire.NewSet(
	// Repository
	deptRepo.NewRepository,
	wire.Bind(new(deptService.Repository), new(*deptRepo.Repository)),

	// Service
	deptService.NewService,

	// Handler
	deptHandler.NewHandler,
)

// App 应用聚合根
type App struct {
	DB           *gorm.DB
	RedisClient  *redis.Client
	TokenManager auth.TokenManager
	Cfg          *config.Config
	DeptHandler  *deptHandler.Handler
}
