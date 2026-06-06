package model

import (
	common "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Role 角色实体
type Role struct {
	ID        types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	Name      string       `gorm:"column:name;not null" json:"name"` // 名称
	Code      string       `gorm:"column:code;not null;uniqueIndex:uk_code" json:"code"` // 编码
	Sort      int          `gorm:"column:sort" json:"sort"` // 排序
	Status    int          `gorm:"column:status;default:1" json:"status"` // 状态(1启用0禁用)
	DataScope int          `gorm:"column:data_scope" json:"dataScope"` // 数据权限范围
	common.BaseEntity
}

func (Role) TableName() string {
	return "sys_role"
}

// RoleMenu 角色菜单关联
type RoleMenu struct {
	RoleID types.BigInt `gorm:"column:role_id;not null;primaryKey" json:"roleId"`
	MenuID types.BigInt `gorm:"column:menu_id;not null;primaryKey" json:"menuId"`
}

func (RoleMenu) TableName() string {
	return "sys_role_menu"
}
