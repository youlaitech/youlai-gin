package main

import (
	"fmt"
	"log"
	"strings"

	"youlai-gin/pkg/config"
)

func main() {
	fmt.Println("=== 多环境配置加载示例 ===\n")

	// 方式 1: 使用默认环境 (dev)
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	printConfig()

	// 方式 3: 使用环境变量 (需提前设置 APP_ENV)
	// Windows: set APP_ENV=prod
	// Linux/Mac: export APP_ENV=prod
	// 然后运行: go run examples/config_example.go
}

func printConfig() {
	cfg := config.Cfg

	fmt.Printf("\n【日志配置】\n")
	fmt.Printf("  级别: %s\n", cfg.Logger.Level)
	fmt.Printf("  控制台输出: %v (颜色: %v)\n",
		cfg.Logger.Console.Enabled, cfg.Logger.Console.Color)
	fmt.Printf("  文件输出: %v (路径: %s)\n",
		cfg.Logger.File.Enabled, cfg.Logger.File.Path)

	fmt.Println("\n" + strings.Repeat("=", 50))
}
