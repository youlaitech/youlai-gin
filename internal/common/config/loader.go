package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load 加载指定环境的配置
func Load(env ...string) error {
	environment := "dev"
	if len(env) > 0 && env[0] != "" {
		environment = env[0]
	} else if envVar := os.Getenv("APP_ENV"); envVar != "" {
		environment = envVar
	}

	configFile := fmt.Sprintf("%s.yaml", environment)
	configPath := filepath.Join("configs", configFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}

	// 校验关键配置项
	if err := validate(&cfg); err != nil {
		return fmt.Errorf("配置校验失败: %w", err)
	}

	Cfg = &cfg

	return nil
}

// validate 校验关键配置项
func validate(cfg *Config) error {
	if cfg.Database.DSN() == "" {
		return fmt.Errorf("数据库 DSN 配置缺失 (database.host/port/user/password/database)")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("Redis 主机配置缺失 (redis.host)")
	}
	if cfg.Redis.Port == 0 {
		return fmt.Errorf("Redis 端口配置缺失 (redis.port)")
	}
	if cfg.Security.JWT.SecretKey == "" {
		return fmt.Errorf("JWT 密钥配置缺失 (security.jwt.secretKey)")
	}
	return nil
}

// GetEnv 获取当前环境
func GetEnv() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	return "dev"
}