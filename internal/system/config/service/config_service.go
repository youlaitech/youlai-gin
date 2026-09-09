package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"youlai-gin/internal/common/redis"
	"youlai-gin/internal/system/config/model"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
)

const (
	configCachePrefix = "sys:config:"
	configCacheExpire = 24 * time.Hour
)

// Repository 配置数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.ConfigQuery) ([]model.Config, int64, error)
	GetByKey(ctx context.Context, configKey string) (*model.Config, error)
	Get(ctx context.Context, id int64) (*model.Config, error)
	Create(ctx context.Context, config *model.Config) error
	Update(ctx context.Context, form *model.ConfigForm) error
	Delete(ctx context.Context, id int64) error
	BatchDelete(ctx context.Context, ids []int64) error
}

// Service 配置业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Page 配置分页列表
func (s *Service) Page(ctx context.Context, query *model.ConfigQuery) (*baseModel.PagedData, error) {
	configs, total, err := s.repo.Page(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询配置列表失败")
	}

	return &baseModel.PagedData{List: configs, Total: total}, nil
}

// GetByKey 根据 Key 获取配置（带缓存）
func (s *Service) GetByKey(ctx context.Context, configKey string) (*model.Config, error) {
	cacheKey := configCachePrefix + configKey
	if cached, err := redis.Client.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var config model.Config
		if err := json.Unmarshal([]byte(cached), &config); err == nil {
			return &config, nil
		}
	}

	config, err := s.repo.GetByKey(ctx, configKey)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(config); err == nil {
		redis.Client.Set(context.Background(), cacheKey, string(data), configCacheExpire)
	}

	return config, nil
}

// Get 根据 ID 获取配置详情
func (s *Service) Get(ctx context.Context, id int64) (*model.Config, error) {
	return s.repo.Get(ctx, id)
}

// GetForm 获取配置表单数据
func (s *Service) GetForm(ctx context.Context, id int64) (*model.ConfigForm, error) {
	config, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errs.NotFound("配置不存在")
	}

	return &model.ConfigForm{
		ID:          config.ID,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		ConfigName:  config.ConfigName,
		Remark:      config.Remark,
	}, nil
}

// Create 新增配置
func (s *Service) Create(ctx context.Context, form *model.ConfigForm) error {
	existing, _ := s.repo.GetByKey(ctx, form.ConfigKey)
	if existing != nil && existing.ID > 0 {
		return errs.Business(fmt.Sprintf("配置Key [%s] 已存在", form.ConfigKey))
	}

	config := &model.Config{
		ConfigKey:   form.ConfigKey,
		ConfigValue: form.ConfigValue,
		ConfigName:  form.ConfigName,
		Remark:      form.Remark,
	}
	if err := s.repo.Create(ctx, config); err != nil {
		return errs.SystemError("新增配置失败")
	}

	s.clearCache(form.ConfigKey)
	return nil
}

// Update 更新配置
func (s *Service) Update(ctx context.Context, form *model.ConfigForm) error {
	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新配置失败")
	}

	s.clearCache(form.ConfigKey)
	return nil
}

// Delete 删除配置（单个或批量）
func (s *Service) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 1 {
		config, err := s.repo.Get(ctx, ids[0])
		if err != nil {
			return errs.NotFound("配置不存在")
		}
		if err := s.repo.Delete(ctx, ids[0]); err != nil {
			return errs.SystemError("删除配置失败")
		}
		s.clearCache(config.ConfigKey)
		return nil
	}

	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		return errs.SystemError("批量删除配置失败")
	}
	s.ClearAllCache(ctx)
	return nil
}

// RefreshCache 刷新指定配置缓存
func (s *Service) RefreshCache(ctx context.Context, configKey string) error {
	s.clearCache(configKey)
	_, err := s.GetByKey(ctx, configKey)
	return err
}

// ClearAllCache 清除所有配置缓存
func (s *Service) ClearAllCache(ctx context.Context) {
	keys, err := redis.Client.Keys(ctx, configCachePrefix+"*").Result()
	if err == nil && len(keys) > 0 {
		redis.Client.Del(ctx, keys...)
	}
}

// clearCache 清除指定配置的缓存（缓存失效用独立上下文，避免请求上下文取消导致残留）
func (s *Service) clearCache(configKey string) {
	redis.Client.Del(context.Background(), configCachePrefix+configKey)
}
