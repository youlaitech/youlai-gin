package model

import "youlai-gin/pkg/types"

// DictPageVO 字典分页数据
type DictPageVO struct {
	ID         types.BigInt `json:"id"` // 主键
	DictCode   string       `json:"dictCode"` // 字典编码
	Name       string       `json:"name"` // 名称
	Status     int          `json:"status"` // 状态(1启用0禁用)
	Remark     string       `json:"remark,omitempty"` // 备注
	CreateTime types.LocalTime `json:"createTime,omitempty"` // 创建时间
	UpdateTime types.LocalTime `json:"updateTime,omitempty"` // 更新时间
}

// DictItemVO 字典项数据
type DictItemVO struct {
	ID     types.BigInt `json:"id"` // 主键
	Value  string       `json:"value"` // 值
	Label  string       `json:"label"` // 标签
	TagType string       `json:"tagType,omitempty"` // 标签类型
	Sort   int          `json:"sort"` // 排序
	Status int          `json:"status"` // 状态(1启用0禁用)
	Remark string       `json:"remark,omitempty"` // 备注
}
