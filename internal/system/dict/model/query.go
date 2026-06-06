package model

import common "youlai-gin/pkg/model"

// DictQuery 字典查询参数
type DictQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"` // 关键字

}

// DictItemQuery 字典项查询参数
type DictItemQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"` // 关键字

	DictCode string `form:"dictCode"` // 字典编码

}
