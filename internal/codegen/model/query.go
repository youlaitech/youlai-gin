package model

import baseModel "youlai-gin/pkg/model"

// TableQuery 数据表分页查询参数
type TableQuery struct {
	baseModel.BaseQuery
	Keywords string `form:"keywords"`
}
