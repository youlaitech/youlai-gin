package main

import (
	"log"

	"youlai-gin/internal/database"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/config"
	"youlai-gin/pkg/logger"
	"youlai-gin/pkg/redis"
)

// 演示完整的初始化流程
func main() {
	// 1. 加载配置（一次性加载所有模块）
	if err := config.Load(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 初始化日志
	logger.InitWithConfig(&config.Cfg.Logger)
	defer logger.Sync()

	logger.Log.Info("开始初始化系统...")

	// 3. 初始化数据库
	if err := database.InitWithConfig(&config.Cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 4. 初始化 Redis
	if err := redis.InitWithConfig(&config.Cfg.Redis); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}
	defer redis.Close()

	// 5. 创建 TokenManager
	tokenManager, err := auth.CreateTokenManager(&config.Cfg.Security)
	if err != nil {
		log.Fatalf("TokenManager 创建失败: %v", err)
	}

	logger.Log.Info("系统初始化完成！")
	logger.Log.Sugar().Infof("会话模式: %s", config.Cfg.Security.SessionType)

	// 6. 测试数据库连接
	sqlDB, _ := database.DB.DB()
	stats := sqlDB.Stats()
	logger.Log.Sugar().Infof("数据库连接池状态: 打开=%d, 空闲=%d", 
		stats.OpenConnections, stats.Idle)

	// 7. 测试 Redis 连接
	logger.Log.Sugar().Infof("Redis 地址: %s:%d", 
		config.Cfg.Redis.Host, config.Cfg.Redis.Port)

	// 8. 测试 TokenManager
	user := &auth.UserDetails{
		UserID:    1,
		Username:  "admin",
		DeptID:    10,
		DataScope: 2,
		Roles:     []string{"ADMIN"},
	}
	token, err := tokenManager.GenerateToken(user)
	if err != nil {
		log.Fatalf("生成 Token 失败: %v", err)
	}
	logger.Log.Sugar().Infof("生成 Token 成功: %s...", token.AccessToken[:50])

	logger.Log.Info("✅ 所有模块测试通过！")
}
