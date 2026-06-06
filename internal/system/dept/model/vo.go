package model

import "youlai-gin/pkg/types"

// DeptVO 部门视图对象
type DeptVO struct {
	ID         types.BigInt `json:"id"` // 主键
	Name       string       `json:"name"` // 名称
	Code       string       `json:"code"` // 编码
	ParentID   types.BigInt `json:"parentId"` // 父级ID
	TreePath   string       `json:"treePath,omitempty"` // 树路径
	Sort       int          `json:"sort"` // 排序
	Status     int          `json:"status"` // 状态(1启用0禁用)
	CreateTime types.LocalTime `json:"createTime,omitempty"` // 创建时间
	UpdateTime types.LocalTime `json:"updateTime,omitempty"` // 更新时间
	Children   []*DeptVO    `json:"children,omitempty"` // 子节点
}

// DeptOption 部门下拉树选项
type DeptOption struct {
	Value    types.BigInt   `json:"value"` // 值
	Label    string         `json:"label"` // 标签
	ParentID types.BigInt   `json:"-"` // 父级ID
	Children []*DeptOption  `json:"children,omitempty"` // 子节点
}
