package model

import "youlai-gin/pkg/types"

// DeptForm 部门表单
type DeptForm struct {
	ID       types.BigInt `json:"id"` // 主键
	Name     string       `json:"name" binding:"required"` // 名称
	Code     string       `json:"code" binding:"required"` // 编码
	ParentID types.BigInt `json:"parentId"` // 父级ID
	Sort     int          `json:"sort"` // 排序
	Status   int          `json:"status" binding:"oneof=0 1"` // 状态(1启用0禁用)
}
