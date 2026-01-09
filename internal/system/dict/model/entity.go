package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

// Dict 字典实体
type Dict struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode string       `gorm:"column:dict_code;uniqueIndex:uk_dict_code" json:"dictCode"`
	Name     string       `gorm:"column:name" json:"name"`
	Status   int          `gorm:"column:status;default:1" json:"status"`
	Remark   string       `gorm:"column:remark" json:"remark"`
	common.BaseEntity
}

// TableName 返回字典表名
func (Dict) TableName() string {
	return "sys_dict"
}

// DictItem 字典项实体
type DictItem struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode string       `gorm:"column:dict_code;index" json:"dictCode"`
	Value    string       `gorm:"column:value" json:"value"`
	Label    string       `gorm:"column:label" json:"label"`
	TagType  string       `gorm:"column:tag_type" json:"tagType"`
	Sort     int          `gorm:"column:sort;default:0" json:"sort"`
	Status   int          `gorm:"column:status;default:1" json:"status"`
	Remark   string       `gorm:"column:remark" json:"remark"`
	common.BaseEntity
}

// TableName 返回字典项表名
func (DictItem) TableName() string {
	return "sys_dict_item"
}
