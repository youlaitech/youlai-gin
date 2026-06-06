package model

import common "youlai-gin/pkg/model"

// Config 系统配置实体
type Config struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	ConfigKey   string `gorm:"column:config_key;size:100;uniqueIndex;not null" json:"configKey"`
	ConfigValue string `gorm:"column:config_value;type:text" json:"configValue"`
	ConfigName  string `gorm:"column:config_name;size:100" json:"configName"`
	Remark      string `gorm:"column:remark;size:500" json:"remark"` // 备注
	common.BaseEntity
}

func (Config) TableName() string {
	return "sys_config"
}
