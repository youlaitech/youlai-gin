package model

// ConfigForm 配置表单
type ConfigForm struct {
	ID          int64  `json:"id"` // 主键
	ConfigKey   string `json:"configKey" binding:"required"`
	ConfigValue string `json:"configValue"`
	ConfigName  string `json:"configName" binding:"required"`
	Remark      string `json:"remark"` // 备注
}
