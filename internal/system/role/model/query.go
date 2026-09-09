package model

import baseModel "youlai-gin/pkg/model"

// RoleQuery 角色查询参数
type RoleQuery struct {
	baseModel.BaseQuery
	Keywords string `form:"keywords"` // 关键字
}
