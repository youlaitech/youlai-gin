package model

import "youlai-gin/pkg/types"

// RolePageVO 角色分页视图
type RolePageVO struct {
	ID            types.BigInt   `json:"id"` // 主键
	Name          string         `json:"name"` // 名称
	Code          string         `json:"code"` // 编码
	Sort          int            `json:"sort"` // 排序
	Status        int            `json:"status"` // 状态(1启用0禁用)
	DataScope     int            `json:"dataScope"` // 数据权限范围
	DataScopeLabel string        `json:"dataScopeLabel"`
	CreateTime    types.LocalTime `json:"createTime"` // 创建时间
	UpdateTime    types.LocalTime `json:"updateTime"` // 更新时间
}
