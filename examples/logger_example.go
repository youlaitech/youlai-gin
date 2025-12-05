package main

import (
	"youlai-gin/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// ========== 方式 1：简单初始化（兼容旧代码）==========
	// logger.Init("dev")  // 开发环境：彩色控制台
	// logger.Init("prod") // 生产环境：JSON + 文件

	// ========== 方式 2：YAML 配置初始化（推荐）==========
	if err := logger.InitFromYAML("config/logger.yaml"); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// ========== 方式 3：代码配置初始化（最灵活）==========
	// cfg := &logger.Config{
	// 	Level: "debug",
	// 	Console: logger.ConsoleConfig{
	// 		Enabled: true,
	// 		Color:   true,
	// 		Format:  "console",
	// 	},
	// 	File: logger.FileConfig{
	// 		Enabled:    true,
	// 		Path:       "logs/app.log",
	// 		ErrorPath:  "logs/error.log",
	// 		Format:     "json",
	// 		MaxSize:    100,
	// 		MaxBackups: 10,
	// 		MaxAge:     30,
	// 		Compress:   true,
	// 	},
	// }
	// logger.InitWithConfig(cfg)

	// ========== 使用日志 ==========
	logger.Log.Info("服务启动",
		zap.String("version", "1.0.0"),
		zap.Int("port", 8000),
	)

	logger.Log.Debug("调试信息", zap.String("key", "value"))
	logger.Log.Warn("警告信息", zap.String("reason", "资源不足"))
	logger.Log.Error("错误信息", zap.Error(nil))

	// ========== 环境变量覆盖示例 ==========
	// export LOG_LEVEL=debug
	// export LOG_COLOR=true
	// export LOG_FILE=true
	// export LOG_FILE_PATH=logs/custom.log
}
