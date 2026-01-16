package config

import (
	"youlai-gin/pkg/ai"
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/logger"
	redisConfig "youlai-gin/pkg/redis"
)

// Config 全局配置
type Config struct {
	Database database.Config     `mapstructure:"database"`
	Logger   logger.Config       `mapstructure:"logger"`
	Redis    redisConfig.Config  `mapstructure:"redis"`
	Security auth.SecurityConfig `mapstructure:"security"`
	AI       ai.Config           `mapstructure:"ai"`
}

// Cfg 全局配置实例
var Cfg *Config
