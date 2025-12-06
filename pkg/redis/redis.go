package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
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

// InitFromYAML 从 YAML 文件初始化 Redis
func InitFromYAML(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取 Redis 配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析 Redis 配置失败: %w", err)
	}

	return InitWithConfig(&cfg)
}

// InitWithConfig 使用配置初始化 Redis
func InitWithConfig(cfg *Config) error {
	Client = redis.NewClient(&redis.Options{
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

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
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
