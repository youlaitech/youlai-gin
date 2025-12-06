package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load 加载指定环境的配置
// 环境优先级：参数 env > 环境变量 APP_ENV > 默认 dev
func Load(env ...string) error {
	// 确定环境
	environment := "dev"
	if len(env) > 0 && env[0] != "" {
		environment = env[0]
	} else if envVar := os.Getenv("APP_ENV"); envVar != "" {
		environment = envVar
	}

	// 配置文件名：dev.yaml / test.yaml / prod.yaml
	configFile := fmt.Sprintf("%s.yaml", environment)
	configPath := filepath.Join("configs", configFile)

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	v := viper.New()

	// 支持环境变量覆盖，前缀为 APP_
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// 设置配置文件
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析到配置结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}

	Cfg = &cfg

	fmt.Printf("✓ 配置加载成功 [环境: %s, 文件: %s]\n", environment, configFile)
	return nil
}

// GetEnv 获取当前环境
func GetEnv() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	return "dev"
}
