package model

import "youlai-gin/pkg/types"

// RoleForm 角色表单
type RoleForm struct {
	ID        types.BigInt   `json:"id"` // 主键
	Name      string         `json:"name" binding:"required"` // 名称
	Code      string         `json:"code" binding:"required"` // 编码
	Sort      int            `json:"sort"` // 排序
	Status    int            `json:"status" binding:"oneof=0 1"` // 状态(1启用0禁用)
	DataScope int            `json:"dataScope" binding:"oneof=1 2 3 4 5"` // 数据权限范围
	DeptIds   []types.BigInt `json:"deptIds"` // 部门ID列表
	MenuIds   []types.BigInt `json:"menuIds" swaggerignore:"true"` // 菜单ID列表
}
