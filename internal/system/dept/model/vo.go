package model

import "youlai-gin/pkg/types"

type DeptVO struct {
	ID         types.BigInt `json:"id"`
	Name       string       `json:"name"`
	Code       string       `json:"code"`
	ParentID   types.BigInt `json:"parentId"`
	TreePath   string       `json:"treePath,omitempty"`
	Sort       int          `json:"sort"`
	Status     int          `json:"status"`
	CreateTime types.LocalTime `json:"createTime,omitempty"`
	UpdateTime types.LocalTime `json:"updateTime,omitempty"`
	Children   []*DeptVO    `json:"children,omitempty"`
}

// DeptOption 部门下拉树选项
type DeptOption struct {
	Value    types.BigInt   `json:"value"`
	Label    string         `json:"label"`
	ParentID types.BigInt   `json:"-"`
	Children []*DeptOption  `json:"children,omitempty"`
}
