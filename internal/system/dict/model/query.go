package model

import "youlai-gin/pkg/common"

type DictPageQuery struct {
	common.PageQuery
	Keywords string `form:"keywords"`
}

type DictItemQuery struct {
	DictCode string `form:"dictCode" binding:"required"`
}
