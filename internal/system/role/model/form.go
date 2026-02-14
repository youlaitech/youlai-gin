package model

import "youlai-gin/pkg/types"

type RoleForm struct {
	ID        types.BigInt   `json:"id"`
	Name      string         `json:"name" binding:"required"`
	Code      string         `json:"code" binding:"required"`
	Sort      int            `json:"sort"`
	Status    int            `json:"status" binding:"oneof=0 1"`
	DataScope int            `json:"dataScope" binding:"oneof=1 2 3 4 5"`
	DeptIds   []types.BigInt `json:"deptIds"`
	MenuIds   []types.BigInt `json:"menuIds" swaggerignore:"true"`
}
