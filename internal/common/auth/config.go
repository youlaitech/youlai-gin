package auth

import (
	"fmt"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	SessionType string           `mapstructure:"sessionType"`
	JWT         JwtConfig        `mapstructure:"jwt"`
	RedisToken  RedisTokenConfig `mapstructure:"redisToken"`
}

// CreateTokenManager 根据配置创建 TokenManager
func CreateTokenManager(cfg *SecurityConfig) (TokenManager, error) {
	switch cfg.SessionType {
	case "jwt":
		return NewJwtTokenManager(&cfg.JWT), nil
	case "redis-token":
		return NewRedisTokenManager(&cfg.RedisToken), nil
	default:
		return nil, fmt.Errorf("不支持的会话类型: %s", cfg.SessionType)
	}
}