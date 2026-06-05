package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 配置
type Config struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	Database int           `yaml:"database"`
	Pool     PoolConfig    `yaml:"pool"`
	Timeout  TimeoutConfig `yaml:"timeout"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdle   int `yaml:"maxIdle"`
	MaxActive int `yaml:"maxActive"`
	MinIdle   int `yaml:"minIdle"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Dial  int `yaml:"dial"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
	Pool  int `yaml:"pool"`
}

var Client *redis.Client

// NewClient Wire Provider：返回 Redis 客户端
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