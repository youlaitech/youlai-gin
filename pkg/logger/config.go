package logger

import (
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Config 日志配置（灵活、可序列化）
type Config struct {
	// 全局配置
	Level      string `yaml:"level" json:"level"`           // 日志级别: debug/info/warn/error
	CallerSkip int    `yaml:"callerSkip" json:"callerSkip"` // 调用栈跳过层数，默认 1

	// 控制台配置
	Console ConsoleConfig `yaml:"console" json:"console"`

	// 文件配置
	File FileConfig `yaml:"file" json:"file"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"` // 是否启用控制台输出
	Color   bool   `yaml:"color" json:"color"`     // 是否启用彩色输出
	Format  string `yaml:"format" json:"format"`   // 格式: json / console
}

// FileConfig 文件输出配置
type FileConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"` // 是否启用文件输出
	Path    string `yaml:"path" json:"path"`       // 日志文件路径（app.log）
	Format  string `yaml:"format" json:"format"`   // 格式: json / console

	// 滚动配置
	MaxSize    int  `yaml:"maxSize" json:"maxSize"`       // 单个文件最大MB
	MaxBackups int  `yaml:"maxBackups" json:"maxBackups"` // 保留旧文件数
	MaxAge     int  `yaml:"maxAge" json:"maxAge"`         // 保留天数
	Compress   bool `yaml:"compress" json:"compress"`     // 是否压缩

	// 按级别分文件
	ErrorPath string `yaml:"errorPath" json:"errorPath"` // 错误日志单独文件（可选）
}

// DefaultConfig 返回默认配置（开发环境友好）
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		CallerSkip: 1,
		Console: ConsoleConfig{
			Enabled: true,
			Color:   true,
			Format:  "console",
		},
		File: FileConfig{
			Enabled:    false,
			Path:       "logs/app.log",
			Format:     "json",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		},
	}
}

// ProductionConfig 返回生产环境配置
func ProductionConfig() *Config {
	return &Config{
		Level:      "info",
		CallerSkip: 1,
		Console: ConsoleConfig{
			Enabled: true,
			Color:   false, // 生产环境可选择关闭彩色
			Format:  "json",
		},
		File: FileConfig{
			Enabled:    true,
			Path:       "/var/log/youlai-gin/app.log",
			ErrorPath:  "/var/log/youlai-gin/error.log", // 错误单独文件
			Format:     "json",
			MaxSize:    100,
			MaxBackups: 30,
			MaxAge:     90,
			Compress:   true,
		},
	}
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// ApplyEnv 从环境变量覆盖配置（12-Factor App）
func (c *Config) ApplyEnv() {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Level = level
	}
	if console := os.Getenv("LOG_CONSOLE"); console == "false" {
		c.Console.Enabled = false
	}
	if color := os.Getenv("LOG_COLOR"); color != "" {
		c.Console.Color = color == "true"
	}
	if file := os.Getenv("LOG_FILE"); file == "true" {
		c.File.Enabled = true
	}
	if path := os.Getenv("LOG_FILE_PATH"); path != "" {
		c.File.Path = path
	}
}
