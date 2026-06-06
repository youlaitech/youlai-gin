package model

import (
	common "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Dept 部门实体
type Dept struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	Name     string       `gorm:"column:name;not null" json:"name"` // 名称
	Code     string       `gorm:"column:code;not null;uniqueIndex:uk_code" json:"code"` // 编码
	ParentID types.BigInt `gorm:"column:parent_id;default:0" json:"parentId"` // 父级ID
	TreePath string       `gorm:"column:tree_path;not null" json:"treePath"` // 树路径
	Sort     int          `gorm:"column:sort;default:0" json:"sort"` // 排序
	Status   int          `gorm:"column:status;default:1" json:"status"` // 状态(1启用0禁用)
	common.BaseEntity
}

func (Dept) TableName() string {
	return "sys_dept"
}
