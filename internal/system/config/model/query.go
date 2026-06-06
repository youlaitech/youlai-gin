package model

import common "youlai-gin/pkg/model"

// ConfigQuery 配置分页查询
type ConfigQuery struct {
	common.BaseQuery
	ConfigKey  string `form:"configKey"` // 配置键

	ConfigName string `form:"configName"` // 配置名称

}

// ConfigListQuery 配置列表查询
type ConfigListQuery struct {
	ConfigKey  string `form:"configKey"` // 配置键

	ConfigName string `form:"configName"` // 配置名称

}
