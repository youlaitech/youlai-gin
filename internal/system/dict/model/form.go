package model

import "youlai-gin/pkg/types"

// DictForm 字典表单
type DictForm struct {
	ID       types.BigInt `json:"id"` // 主键
	DictCode string       `json:"dictCode" binding:"required"` // 字典编码
	Name     string       `json:"name" binding:"required"` // 名称
	Status   int          `json:"status" binding:"oneof=0 1"` // 状态(1启用0禁用)
	Remark   string       `json:"remark"` // 备注
}

// DictItemForm 字典项表单
type DictItemForm struct {
	ID       types.BigInt `json:"id"` // 主键
	DictCode string       `json:"dictCode" binding:"required"` // 字典编码
	Value    string       `json:"value" binding:"required"` // 值
	Label    string       `json:"label" binding:"required"` // 标签
	TagType  string       `json:"tagType"` // 标签类型
	Sort     int          `json:"sort"` // 排序
	Status   int          `json:"status" binding:"oneof=0 1"` // 状态(1启用0禁用)
	Remark   string       `json:"remark"` // 备注
}
