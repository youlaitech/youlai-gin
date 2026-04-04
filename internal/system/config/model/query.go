package model

import common "youlai-gin/pkg/model"

// ConfigQuery 配置分页查询
type ConfigQuery struct {
	common.BaseQuery
	ConfigKey  string `form:"configKey"`
	ConfigName string `form:"configName"`
}

// ConfigListQuery 配置列表查询
type ConfigListQuery struct {
	ConfigKey  string `form:"configKey"`
	ConfigName string `form:"configName"`
}
