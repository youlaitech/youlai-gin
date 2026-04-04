package model

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
