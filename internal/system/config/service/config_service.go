package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	
	"youlai-gin/internal/system/config/model"
	"youlai-gin/internal/system/config/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
)

const (
	configCachePrefix = "sys:config:"
	configCacheExpire = 24 * time.Hour
)

// GetConfigList 获取配置列表
func GetConfigList(query *model.ConfigQuery) ([]model.Config, error) {
	return repository.GetConfigList(query)
}

// GetConfigPage 获取配置分页列表
func GetConfigPage(query *model.ConfigPageQuery) (*common.PageResult, error) {
	configs, total, err := repository.GetConfigPage(query)
	if err != nil {
		return nil, errs.SystemError("查询配置列表失败")
	}
	
	return &common.PageResult{
		List:  configs,
		Total: total,
	}, nil
}

// GetAllConfigs 获取所有配置
func GetAllConfigs() ([]model.Config, error) {
	return repository.GetConfigList(&model.ConfigQuery{})
}

// GetConfigByKey 根据Key获取配置（带缓存）
func GetConfigByKey(configKey string) (*model.Config, error) {
	// 先从缓存获取
	cacheKey := configCachePrefix + configKey
	cached, err := redis.Client.Get(context.Background(), cacheKey).Result()
	
	if err == nil && cached != "" {
		var config model.Config
		if err := json.Unmarshal([]byte(cached), &config); err == nil {
			return &config, nil
		}
	}
	
	// 缓存未命中，从数据库查询
	config, err := repository.GetConfigByKey(configKey)
	if err != nil {
		return nil, err
	}
	
	// 写入缓存
	if data, err := json.Marshal(config); err == nil {
		redis.Client.Set(context.Background(), cacheKey, string(data), configCacheExpire)
	}
	
	return config, nil
}

// GetConfigValue 获取配置值（字符串）
func GetConfigValue(configKey string) (string, error) {
	config, err := GetConfigByKey(configKey)
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

// GetConfigValueWithDefault 获取配置值，如果不存在返回默认值
func GetConfigValueWithDefault(configKey, defaultValue string) string {
	value, err := GetConfigValue(configKey)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetConfigInt 获取配置值（整数）
func GetConfigInt(configKey string) (int, error) {
	value, err := GetConfigValue(configKey)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetConfigBool 获取配置值（布尔）
func GetConfigBool(configKey string) (bool, error) {
	value, err := GetConfigValue(configKey)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// GetConfigByID 根据ID获取配置
func GetConfigByID(id int64) (*model.Config, error) {
	return repository.GetConfigByID(id)
}

// GetConfigFormData 获取配置表单数据
func GetConfigFormData(id int64) (*model.ConfigForm, error) {
	config, err := repository.GetConfigByID(id)
	if err != nil {
		return nil, errs.NotFound("配置不存在")
	}
	
	return &model.ConfigForm{
		ID:          config.ID,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		ConfigName:  config.ConfigName,
		ConfigType:  config.ConfigType,
		Description: config.Description,
		Sort:        config.Sort,
	}, nil
}

// SaveConfig 保存配置（新增或更新）
func SaveConfig(form *model.ConfigForm) error {
	config := &model.Config{
		ID:          form.ID,
		ConfigKey:   form.ConfigKey,
		ConfigValue: form.ConfigValue,
		ConfigName:  form.ConfigName,
		ConfigType:  form.ConfigType,
		Description: form.Description,
		Sort:        form.Sort,
	}
	
	var err error
	if config.ID > 0 {
		// 更新
		err = repository.UpdateConfig(config)
	} else {
		// 新增 - 检查Key是否已存在
		existing, _ := repository.GetConfigByKey(config.ConfigKey)
		if existing != nil && existing.ID > 0 {
			return errs.BadRequest(fmt.Sprintf("配置Key [%s] 已存在", config.ConfigKey))
		}
		err = repository.CreateConfig(config)
	}
	
	if err != nil {
		return errs.SystemError("保存配置失败")
	}
	
	// 清除缓存
	ClearConfigCache(config.ConfigKey)
	
	return nil
}

// DeleteConfig 删除配置
func DeleteConfig(id int64) error {
	config, err := repository.GetConfigByID(id)
	if err != nil {
		return errs.NotFound("配置不存在")
	}
	
	if err := repository.DeleteConfig(id); err != nil {
		return errs.SystemError("删除配置失败")
	}
	
	// 清除缓存
	ClearConfigCache(config.ConfigKey)
	
	return nil
}

// BatchDeleteConfig 批量删除配置
func BatchDeleteConfig(ids []int64) error {
	if err := repository.BatchDeleteConfig(ids); err != nil {
		return errs.SystemError("批量删除配置失败")
	}
	
	// 清除所有配置缓存
	ClearAllConfigCache()
	
	return nil
}

// ClearConfigCache 清除指定配置的缓存
func ClearConfigCache(configKey string) {
	cacheKey := configCachePrefix + configKey
	redis.Client.Del(context.Background(), cacheKey)
}

// ClearAllConfigCache 清除所有配置缓存
func ClearAllConfigCache() {
	ctx := context.Background()
	keys, err := redis.Client.Keys(ctx, configCachePrefix+"*").Result()
	if err == nil && len(keys) > 0 {
		redis.Client.Del(ctx, keys...)
	}
}

// RefreshConfigCache 刷新配置缓存
func RefreshConfigCache(configKey string) error {
	ClearConfigCache(configKey)
	_, err := GetConfigByKey(configKey)
	return err
}
