package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

// Role 角色实体
type Role struct {
	ID        types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string       `gorm:"column:name;not null" json:"name"`
	Code      string       `gorm:"column:code;not null;uniqueIndex:uk_code" json:"code"`
	Sort      int          `gorm:"column:sort" json:"sort"`
	Status    int          `gorm:"column:status;default:1" json:"status"`
	DataScope int          `gorm:"column:data_scope" json:"dataScope"`
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
