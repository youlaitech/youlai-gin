package model

import common "youlai-gin/pkg/model"

type RoleQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"`
}
