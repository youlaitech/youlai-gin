package model

import "youlai-gin/pkg/common"

// DictQuery 字典查询参数
type DictQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"`
}

// DictItemQuery 字典项查询参数
type DictItemQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"`
	DictCode string `form:"dictCode"`
}
