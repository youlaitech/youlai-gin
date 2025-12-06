package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitWithConfig 使用配置初始化数据库
func InitWithConfig(cfg *Config) error {
	dsn := cfg.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 应用连接池配置
	if err := cfg.ApplyConnectionPool(db); err != nil {
		return fmt.Errorf("配置连接池失败: %w", err)
	}

	DB = db
	log.Println("✓ 数据库连接成功")
	return nil
}
