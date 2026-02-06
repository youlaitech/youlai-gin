package database

import (
	"gorm.io/gorm"
)

// PaginateConfig 分页配置
type PaginateConfig struct {
	Page         int // 页码，从 1 开始
	PageSize     int // 每页大小
	MaxPageSize  int // 最大每页大小，0 表示不限制
	DefaultSize  int // 默认每页大小
}

// DefaultPaginateConfig 默认分页配置
var DefaultPaginateConfig = PaginateConfig{
	Page:        1,
	PageSize:    10,
	MaxPageSize: 100,
	DefaultSize: 10,
}

// Paginate 通用分页函数
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return PaginateWithConfig(page, pageSize, DefaultPaginateConfig)
}

// PaginateWithConfig 带配置的分页函数
func PaginateWithConfig(page, pageSize int, config PaginateConfig) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 页码校验
		if page <= 0 {
			page = config.Page
			if page <= 0 {
				page = 1
			}
		}

		// 每页大小校验
		if pageSize <= 0 {
			pageSize = config.DefaultSize
			if pageSize <= 0 {
				pageSize = 10
			}
		}

		// 最大每页大小限制
		if config.MaxPageSize > 0 && pageSize > config.MaxPageSize {
			pageSize = config.MaxPageSize
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// PaginateQuery 从查询参数构建分页
type PaginateQuery interface {
	GetPage() int
	GetPageSize() int
}

// PaginateFromQuery 从查询对象构建分页
func PaginateFromQuery(query PaginateQuery) func(db *gorm.DB) *gorm.DB {
	return Paginate(query.GetPage(), query.GetPageSize())
}
