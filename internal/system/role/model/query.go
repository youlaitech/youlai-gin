package model

import "youlai-gin/pkg/common"

type RolePageQuery struct {
	common.PageQuery
	Keywords string `form:"keywords"`
}
