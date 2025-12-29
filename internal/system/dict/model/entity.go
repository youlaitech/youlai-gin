package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

type Dict struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode string       `gorm:"column:dict_code;uniqueIndex:uk_dict_code" json:"dictCode"`
	Name     string       `gorm:"column:name" json:"name"`
	Status   int          `gorm:"column:status;default:1" json:"status"`
	Remark   string       `gorm:"column:remark" json:"remark"`
	common.BaseEntity
}

func (Dict) TableName() string {
	return "sys_dict"
}

type DictItem struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode string       `gorm:"column:dict_code;index" json:"dictCode"`
	Value    string       `gorm:"column:value" json:"value"`
	Label    string       `gorm:"column:label" json:"label"`
	Sort     int          `gorm:"column:sort;default:0" json:"sort"`
	Status   int          `gorm:"column:status;default:1" json:"status"`
	Remark   string       `gorm:"column:remark" json:"remark"`
	common.BaseEntity
}

func (DictItem) TableName() string {
	return "sys_dict_item"
}
