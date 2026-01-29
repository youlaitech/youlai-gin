package common

import "youlai-gin/pkg/types"

// BaseEntity 基础实体（所有表的公共字段）
type BaseEntity struct {
	CreateBy   *types.BigInt `gorm:"column:create_by" json:"createBy,omitempty"`
	CreateTime types.LocalTime `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
	UpdateBy   *types.BigInt `gorm:"column:update_by" json:"updateBy,omitempty"`
	UpdateTime types.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	IsDeleted  int           `gorm:"column:is_deleted;default:0" json:"-"`
}
