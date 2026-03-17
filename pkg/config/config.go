package config

import (
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/logger"
	redisConfig "youlai-gin/pkg/redis"
)

// WechatConfig 微信配置
type WechatConfig struct {
	Miniapp struct {
		AppID     string `mapstructure:"appId"`
		AppSecret string `mapstructure:"appSecret"`
	} `mapstructure:"miniapp"`
}

// Config 全局配置
type Config struct {
	Database database.Config     `mapstructure:"database"`
	Logger   logger.Config       `mapstructure:"logger"`
	Redis    redisConfig.Config  `mapstructure:"redis"`
	Security auth.SecurityConfig `mapstructure:"security"`
	Wechat   WechatConfig        `mapstructure:"wechat"`
}

// Cfg 全局配置实例
var Cfg *Config
