package model

import (
	"time"

	"youlai-gin/pkg/types"
)

// Menu 菜单实体
type Menu struct {
	ID         types.BigInt     `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	ParentID   types.BigInt     `gorm:"column:parent_id;not null" json:"parentId"` // 父级ID
	TreePath   string           `gorm:"column:tree_path" json:"treePath"` // 树路径
	Name       string           `gorm:"column:name;not null" json:"name"` // 名称
	Type       string           `gorm:"column:type;not null" json:"type"` // 类型(C目录/M菜单/B按钮)
	RouteName  string           `gorm:"column:route_name" json:"routeName"` // 路由名称
	RoutePath  string           `gorm:"column:route_path" json:"routePath"` // 路由路径
	Component  string           `gorm:"column:component" json:"component"` // 前端组件
	Perm       string           `gorm:"column:perm" json:"perm"` // 权限标识
	AlwaysShow int              `gorm:"column:always_show;default:0" json:"alwaysShow"` // 始终显示
	KeepAlive  int              `gorm:"column:keep_alive;default:0" json:"keepAlive"` // 页面缓存
	Visible    int              `gorm:"column:visible;default:1" json:"visible"` // 是否可见
	Sort       int              `gorm:"column:sort;default:0" json:"sort"` // 排序
	Icon       string           `gorm:"column:icon" json:"icon"` // 图标
	Redirect   string           `gorm:"column:redirect" json:"redirect"` // 重定向
	CreateTime time.Time        `gorm:"column:create_time;autoCreateTime" json:"createTime"` // 创建时间
	UpdateTime time.Time        `gorm:"column:update_time;autoUpdateTime" json:"updateTime"` // 更新时间
	Params     map[string]any   `gorm:"column:params;serializer:json" json:"params"` // 路由参数
}

func (Menu) TableName() string {
	return "sys_menu"
}
