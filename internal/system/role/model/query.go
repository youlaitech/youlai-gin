package model

import common "youlai-gin/pkg/model"

// RoleQuery 角色查询参数
type RoleQuery struct {
	common.BaseQuery
	Keywords string `form:"keywords"` // 关键字

}
