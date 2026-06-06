package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 配置
type Config struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Password string        `mapstructure:"password"`
	Database int           `mapstructure:"database"`
	Pool     PoolConfig    `mapstructure:"pool"`
	Timeout  TimeoutConfig `mapstructure:"timeout"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdle   int `mapstructure:"maxIdle"`
	MaxActive int `mapstructure:"maxActive"`
	MinIdle   int `mapstructure:"minIdle"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Dial  int `mapstructure:"dial"`
	Read  int `mapstructure:"read"`
	Write int `mapstructure:"write"`
	Pool  int `mapstructure:"pool"`
}

var Client *redis.Client

// NewClient 创建 Redis 客户端
func NewClient(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.Pool.MaxActive,
		MinIdleConns: cfg.Pool.MinIdle,
		DialTimeout:  time.Duration(cfg.Timeout.Dial) * time.Second,
		ReadTimeout:  time.Duration(cfg.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.Timeout.Write) * time.Second,
		PoolTimeout:  time.Duration(cfg.Timeout.Pool) * time.Second,
	})

	Client = client
	return client
}

// InitWithConfig 使用配置初始化 Redis（main.go 启动调用）
func InitWithConfig(cfg *Config) error {
	client := NewClient(cfg)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	return nil
}

// Close 关闭 Redis 连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}