package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// InitWithConfig 使用配置初始化日志
func InitWithConfig(cfg *Config) {

	// 基础编码配置
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder

	var cores []zapcore.Core

	// 1. 控制台输出
	if cfg.Console.Enabled {
		consoleEncoder := buildEncoder(encoderCfg, cfg.Console.Format, cfg.Console.Color)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			parseLevel(cfg.Level),
		)
		cores = append(cores, consoleCore)
	}

	// 2. 文件输出（所有日志）
	if cfg.File.Enabled && cfg.File.Path != "" {
		// 确保目录存在
		dir := filepath.Dir(cfg.File.Path)
		_ = os.MkdirAll(dir, 0755)

		fileEncoder := buildEncoder(encoderCfg, cfg.File.Format, false)
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   cfg.File.Compress,
		})
		fileCore := zapcore.NewCore(
			fileEncoder,
			fileWriter,
			parseLevel(cfg.Level),
		)
		cores = append(cores, fileCore)
	}

	// 3. 错误日志单独文件（可选）
	if cfg.File.Enabled && cfg.File.ErrorPath != "" {
		dir := filepath.Dir(cfg.File.ErrorPath)
		_ = os.MkdirAll(dir, 0755)

		errorEncoder := buildEncoder(encoderCfg, cfg.File.Format, false)
		errorWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File.ErrorPath,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   cfg.File.Compress,
		})
		// 只记录 Error 及以上级别
		errorCore := zapcore.NewCore(
			errorEncoder,
			errorWriter,
			zapcore.ErrorLevel,
		)
		cores = append(cores, errorCore)
	}

	// 合并所有 Core
	core := zapcore.NewTee(cores...)

	// 构建 Logger
	options := []zap.Option{
		zap.AddCaller(),
	}
	if cfg.CallerSkip > 0 {
		options = append(options, zap.AddCallerSkip(cfg.CallerSkip))
	}

	logger := zap.New(core, options...)
	Log = logger
	zap.ReplaceGlobals(logger)
}

// buildEncoder 构建编码器
func buildEncoder(base zapcore.EncoderConfig, format string, color bool) zapcore.Encoder {
	cfg := base
	if color {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if format == "console" {
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

// Sync 刷盘
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Info 记录Info级别日志
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Error 记录Error级别日志
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Debug 记录Debug级别日志
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Warn 记录Warn级别日志
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

