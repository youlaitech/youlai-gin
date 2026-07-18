package config

import (
	"strconv"
	"strings"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/logger"
	redisConfig "youlai-gin/internal/common/redis"
)

// WechatConfig 微信配置
type WechatConfig struct {
	Miniapp struct {
		AppID     string `mapstructure:"appId"`
		AppSecret string `mapstructure:"appSecret"`
	} `mapstructure:"miniapp"`
}

// FileStorageConfig 文件存储配置（结构参照 youlai-boot 的 file-storage）
type FileStorageConfig struct {
	Type   string                  `mapstructure:"type"`   // 存储类型：minio | aliyun | local
	Upload FileStorageUploadConfig `mapstructure:"upload"` // 上传限制
	Minio  FileStorageMinioConfig  `mapstructure:"minio"`  // MinIO 对象存储
	Local  FileStorageLocalConfig  `mapstructure:"local"`  // 本地存储
}

// FileStorageUploadConfig 上传限制
type FileStorageUploadConfig struct {
	MaxFileSize       string   `mapstructure:"max-file-size"`      // 单文件大小上限（DataSize 字符串，如 50MB）
	AllowedExtensions []string `mapstructure:"allowed-extensions"` // 允许的文件扩展名白名单（置空表示不限制）
}

// FileStorageMinioConfig MinIO 对象存储
type FileStorageMinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`   // 服务地址（含 scheme，如 http://host:9000）
	AccessKey string `mapstructure:"access-key"` // 访问凭据
	SecretKey string `mapstructure:"secret-key"` // 凭据密钥
	Bucket    string `mapstructure:"bucket"`     // 存储桶名称
	Domain    string `mapstructure:"domain"`     // 自定义域名（配置后 URL 走域名，留空用 endpoint）
}

// FileStorageLocalConfig 本地存储
type FileStorageLocalConfig struct {
	Path    string `mapstructure:"path"`     // 存储根目录
	BaseURL string `mapstructure:"base-url"` // 访问基础路径（拼接存储路径即为可访问 URL）
}

// MaxFileSizeBytes 将 max-file-size 字符串解析为字节数（无法解析时返回 0，表示不限制）
func (u FileStorageUploadConfig) MaxFileSizeBytes() int64 {
	return ParseFileSize(u.MaxFileSize)
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	DefaultLimit  int               `mapstructure:"default-limit"`
	DefaultWindow string            `mapstructure:"default-window"`
	Ip            IpRateLimitConfig `mapstructure:"ip"`
}

// IpRateLimitConfig IP 全局限流配置
type IpRateLimitConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用
	Limit   int    `mapstructure:"limit"`   // 窗口内最大请求数
	Window  string `mapstructure:"window"`  // 滑动窗口（如 60s）
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// Config 全局配置
type Config struct {
	Server      ServerConfig        `mapstructure:"server"`
	Database    database.Config     `mapstructure:"database"`
	Logger      logger.Config       `mapstructure:"logger"`
	Redis       redisConfig.Config  `mapstructure:"redis"`
	Security    auth.SecurityConfig `mapstructure:"security"`
	Wechat      WechatConfig        `mapstructure:"wechat"`
	RateLimit   RateLimitConfig     `mapstructure:"rate-limit"`
	FileStorage FileStorageConfig   `mapstructure:"file-storage"`
}

// Cfg 全局配置实例
var Cfg *Config

// ParseFileSize 解析 DataSize 字符串为字节数（支持 B/KB/MB/GB，可省略单位视为字节）
func ParseFileSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	multiplier := int64(1)
	upper := strings.ToUpper(s)
	switch {
	case strings.HasSuffix(upper, "GB"):
		multiplier = 1 << 30
		s = s[:len(s)-2]
	case strings.HasSuffix(upper, "MB"):
		multiplier = 1 << 20
		s = s[:len(s)-2]
	case strings.HasSuffix(upper, "KB"):
		multiplier = 1 << 10
		s = s[:len(s)-2]
	case strings.HasSuffix(upper, "B"):
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return n * multiplier
}
