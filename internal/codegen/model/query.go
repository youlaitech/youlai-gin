package model

import common "youlai-gin/pkg/model"

// TableQuery 数据表分页查询参数
type TableQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"`
}
