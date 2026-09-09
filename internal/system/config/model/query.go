package model

import baseModel "youlai-gin/pkg/model"

// ConfigQuery 配置分页查询
type ConfigQuery struct {
	baseModel.BaseQuery
	ConfigKey  string `form:"configKey"`  // 配置键
	ConfigName string `form:"configName"` // 配置名称
}
