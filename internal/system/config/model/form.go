package model

import "youlai-gin/pkg/common"

// ConfigForm 配置表单
type ConfigForm struct {
	ID          int64  `json:"id"`
	ConfigKey   string `json:"configKey" binding:"required"`
	ConfigValue string `json:"configValue"`
	ConfigName  string `json:"configName" binding:"required"`
	ConfigType  string `json:"configType"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

// ConfigQuery 配置查询
type ConfigListQuery struct {
	ConfigKey  string `form:"configKey"`
	ConfigName string `form:"configName"`
}

// ConfigPageQuery 配置分页查询
type ConfigQuery struct {
	common.BaseQuery
	ConfigKey  string `form:"configKey"`
	ConfigName string `form:"configName"`
}
