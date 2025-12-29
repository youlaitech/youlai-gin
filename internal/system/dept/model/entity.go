package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

type Dept struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string       `gorm:"column:name;not null" json:"name"`
	Code     string       `gorm:"column:code;not null;uniqueIndex:uk_code" json:"code"`
	ParentID types.BigInt `gorm:"column:parent_id;default:0" json:"parentId"`
	TreePath string       `gorm:"column:tree_path;not null" json:"treePath"`
	Sort     int          `gorm:"column:sort;default:0" json:"sort"`
	Status   int          `gorm:"column:status;default:1" json:"status"`
	common.BaseEntity
}

func (Dept) TableName() string {
	return "sys_dept"
}
