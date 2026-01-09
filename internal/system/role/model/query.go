package model

import "youlai-gin/pkg/common"

type RoleQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"`
}
