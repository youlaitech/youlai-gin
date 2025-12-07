package model

import "youlai-gin/pkg/common"

// Config 系统配置实体
type Config struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string `gorm:"column:config_key;size:100;uniqueIndex;not null" json:"configKey"`
	ConfigValue string `gorm:"column:config_value;type:text" json:"configValue"`
	ConfigName  string `gorm:"column:config_name;size:100" json:"configName"`
	ConfigType  string `gorm:"column:config_type;size:20;default:text" json:"configType"` // text, number, boolean, json
	Description string `gorm:"column:description;size:500" json:"description"`
	Sort        int    `gorm:"column:sort;default:0" json:"sort"`
	common.BaseEntity
}

func (Config) TableName() string {
	return "sys_config"
}
