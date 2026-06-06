package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"youlai-gin/internal/common/logger"
)

var DB *gorm.DB

// NewDB 创建数据库连接
func NewDB(cfg *Config) (*gorm.DB, error) {
	dsn := cfg.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := cfg.ApplyConnectionPool(db); err != nil {
		return nil, fmt.Errorf("配置连接池失败: %w", err)
	}

	DB = db
	logger.Log.Sugar().Info("数据库连接成功")
	return db, nil
}

// InitWithConfig 使用配置初始化数据库（main.go 启动调用）
func InitWithConfig(cfg *Config) error {
	_, err := NewDB(cfg)
	return err
}