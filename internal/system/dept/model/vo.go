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
	CreateTime string       `json:"createTime,omitempty"`
	UpdateTime string       `json:"updateTime,omitempty"`
	Children   []*DeptVO    `json:"children,omitempty"`
}
