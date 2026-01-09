package repository

import (
	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/config/model"
	pkgDatabase "youlai-gin/pkg/database"
)

// GetConfigList 获取配置列表
func GetConfigList(query *model.ConfigListQuery) ([]model.Config, error) {
	var configs []model.Config
	db := database.DB.Model(&model.Config{}).Where("is_deleted = 0")

	if query.ConfigKey != "" {
		db = db.Where("config_key LIKE ?", "%"+query.ConfigKey+"%")
	}

	if query.ConfigName != "" {
		db = db.Where("config_name LIKE ?", "%"+query.ConfigName+"%")
	}

	err := db.Order("sort ASC, id ASC").Find(&configs).Error
	return configs, err
}

// GetConfigPage 获取配置分页列表
func GetConfigPage(query *model.ConfigQuery) ([]model.Config, int64, error) {
	var configs []model.Config
	var total int64

	db := database.DB.Model(&model.Config{}).Where("is_deleted = 0")

	if query.ConfigKey != "" {
		db = db.Where("config_key LIKE ?", "%"+query.ConfigKey+"%")
	}

	if query.ConfigName != "" {
		db = db.Where("config_name LIKE ?", "%"+query.ConfigName+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("id ASC").
		Find(&configs).Error

	return configs, total, err
}

// GetConfigByKey 根据Key获取配置
func GetConfigByKey(configKey string) (*model.Config, error) {
	var config model.Config
	err := database.DB.Where("config_key = ? AND is_deleted = 0", configKey).First(&config).Error
	return &config, err
}

// GetConfigByID 根据ID获取配置
func GetConfigByID(id int64) (*model.Config, error) {
	var config model.Config
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&config).Error
	return &config, err
}

// CreateConfig 创建配置
func CreateConfig(config *model.Config) error {
	return database.DB.Create(config).Error
}

// UpdateConfig 更新配置
func UpdateConfig(config *model.Config) error {
	return database.DB.Model(config).Updates(config).Error
}

// DeleteConfig 删除配置
func DeleteConfig(id int64) error {
	return database.DB.Model(&model.Config{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// BatchDeleteConfig 批量删除配置
func BatchDeleteConfig(ids []int64) error {
	return database.DB.Model(&model.Config{}).Where("id IN ?", ids).Update("is_deleted", 1).Error
}
