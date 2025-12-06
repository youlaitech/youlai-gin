package auth

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	SessionType string             `yaml:"sessionType"`
	JWT         JwtConfig          `yaml:"jwt"`
	RedisToken  RedisTokenConfig   `yaml:"redisToken"`
}

// LoadSecurityConfig 从 YAML 加载安全配置
func LoadSecurityConfig(path string) (*SecurityConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取安全配置文件失败: %w", err)
	}

	var cfg SecurityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析安全配置失败: %w", err)
	}

	return &cfg, nil
}

// CreateTokenManager 根据配置创建 TokenManager
func CreateTokenManager(cfg *SecurityConfig) (TokenManager, error) {
	switch cfg.SessionType {
	case "jwt":
		return NewJwtTokenManager(&cfg.JWT), nil
	case "redis-token":
		return NewRedisTokenManager(&cfg.RedisToken), nil
	default:
		return nil, fmt.Errorf("unsupported session type: %s", cfg.SessionType)
	}
}
