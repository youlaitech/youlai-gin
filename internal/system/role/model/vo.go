package model

import "youlai-gin/pkg/types"

type RolePageVO struct {
	ID         types.BigInt `json:"id"`
	Name       string       `json:"name"`
	Code       string       `json:"code"`
	Sort       int          `json:"sort"`
	Status     int          `json:"status"`
	DataScope  int          `json:"dataScope"`
	CreateTime types.LocalTime `json:"createTime"`
	UpdateTime types.LocalTime `json:"updateTime"`
}
