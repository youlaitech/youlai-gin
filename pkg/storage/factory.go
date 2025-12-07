package storage

import (
	"fmt"
)

// NewStorage 创建存储实例（工厂函数）
func NewStorage(config *Config) (Storage, error) {
	switch config.Type {
	case TypeLocal:
		return NewLocalStorage(config.BasePath, config.Domain), nil
		
	case TypeAliyun:
		return NewAliyunOSS(config)
		
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", config.Type)
	}
}

// DefaultStorage 全局默认存储实例
var DefaultStorage Storage

// InitDefaultStorage 初始化默认存储
func InitDefaultStorage(config *Config) error {
	storage, err := NewStorage(config)
	if err != nil {
		return err
	}
	DefaultStorage = storage
	return nil
}
