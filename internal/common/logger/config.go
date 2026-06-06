package logger

import (
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Config 日志配置
type Config struct {
	Level      string        `mapstructure:"level"`      // 日志级别: debug/info/warn/error
	CallerSkip int           `mapstructure:"callerSkip"` // 调用栈跳过层数，默认 1
	Console    ConsoleConfig `mapstructure:"console"`
	File       FileConfig    `mapstructure:"file"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用控制台输出
	Color   bool   `mapstructure:"color"`   // 是否启用彩色输出
	Format  string `mapstructure:"format"`  // 格式: json / console
}

// FileConfig 文件输出配置
type FileConfig struct {
	Enabled    bool   `mapstructure:"enabled"`    // 是否启用文件输出
	Path       string `mapstructure:"path"`       // 日志文件路径（app.log）
	Format     string `mapstructure:"format"`     // 格式: json / console
	MaxSize    int    `mapstructure:"maxSize"`    // 单个文件最大MB
	MaxBackups int    `mapstructure:"maxBackups"` // 保留旧文件数
	MaxAge     int    `mapstructure:"maxAge"`     // 保留天数
	Compress   bool   `mapstructure:"compress"`   // 是否压缩
	ErrorPath  string `mapstructure:"errorPath"`  // 错误日志单独文件（可选）
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