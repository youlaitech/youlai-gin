package model

import "youlai-gin/pkg/types"

type DictPageVO struct {
	ID         types.BigInt `json:"id"`
	DictCode   string       `json:"dictCode"`
	Name       string       `json:"name"`
	Status     int          `json:"status"`
	Remark     string       `json:"remark,omitempty"`
	CreateTime string       `json:"createTime,omitempty"`
	UpdateTime string       `json:"updateTime,omitempty"`
}

type DictItemVO struct {
	ID     types.BigInt `json:"id"`
	Value  string       `json:"value"`
	Label  string       `json:"label"`
	Sort   int          `json:"sort"`
	Status int          `json:"status"`
	Remark string       `json:"remark,omitempty"`
}
