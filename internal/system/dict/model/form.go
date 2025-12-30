package model

import "youlai-gin/pkg/types"

type DictForm struct {
	ID       types.BigInt `json:"id"`
	DictCode string       `json:"dictCode" binding:"required"`
	Name     string       `json:"name" binding:"required"`
	Status   int          `json:"status" binding:"oneof=0 1"`
	Remark   string       `json:"remark"`
}

type DictItemForm struct {
	ID       types.BigInt `json:"id"`
	DictCode string       `json:"dictCode" binding:"required"`
	Value    string       `json:"value" binding:"required"`
	Label    string       `json:"label" binding:"required"`
	Sort     int          `json:"sort"`
	Status   int          `json:"status" binding:"oneof=0 1"`
	Remark   string       `json:"remark"`
}
