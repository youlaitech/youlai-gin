package model

import "youlai-gin/pkg/types"

type DeptForm struct {
	ID       types.BigInt `json:"id"`
	Name     string       `json:"name" binding:"required"`
	Code     string       `json:"code" binding:"required"`
	ParentID types.BigInt `json:"parentId"`
	Sort     int          `json:"sort"`
	Status   int          `json:"status" binding:"oneof=0 1"`
}
