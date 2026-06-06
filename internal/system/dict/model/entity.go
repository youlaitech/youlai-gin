package model

import (
	common "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Dict 字典实体
type Dict struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	DictCode string       `gorm:"column:dict_code;uniqueIndex:uk_dict_code" json:"dictCode"` // 字典编码
	Name     string       `gorm:"column:name" json:"name"` // 名称
	Status   int          `gorm:"column:status;default:1" json:"status"` // 状态(1启用0禁用)
	Remark   string       `gorm:"column:remark" json:"remark"` // 备注
	common.BaseEntity
}

// TableName 返回字典表名
func (Dict) TableName() string {
	return "sys_dict"
}

// DictItem 字典项实体
type DictItem struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	DictCode string       `gorm:"column:dict_code;index" json:"dictCode"` // 字典编码
	Value    string       `gorm:"column:value" json:"value"` // 值
	Label    string       `gorm:"column:label" json:"label"` // 标签
	TagType  string       `gorm:"column:tag_type" json:"tagType"` // 标签类型
	Sort     int          `gorm:"column:sort;default:0" json:"sort"` // 排序
	Status   int          `gorm:"column:status;default:1" json:"status"` // 状态(1启用0禁用)
	Remark   string       `gorm:"column:remark" json:"remark"` // 备注
	CreateBy   *types.BigInt    `gorm:"column:create_by" json:"createBy,omitempty"`
	CreateTime types.LocalTime  `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"` // 创建时间
	UpdateBy   *types.BigInt    `gorm:"column:update_by" json:"updateBy,omitempty"`
	UpdateTime types.LocalTime  `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"` // 更新时间
}

// TableName 返回字典项表名
func (DictItem) TableName() string {
	return "sys_dict_item"
}
