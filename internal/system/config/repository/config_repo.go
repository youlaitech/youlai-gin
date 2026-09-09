package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/system/config/model"
	"youlai-gin/pkg/gormx"
)

// Repository 配置数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Page 配置分页查询
func (r *Repository) Page(ctx context.Context, query *model.ConfigQuery) ([]model.Config, int64, error) {
	var configs []model.Config
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Config{}).Where("is_deleted = 0")

	if query.ConfigKey != "" {
		db = db.Where("config_key LIKE ?", "%"+query.ConfigKey+"%")
	}

	if query.ConfigName != "" {
		db = db.Where("config_name LIKE ?", "%"+query.ConfigName+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("id ASC").
		Find(&configs).Error

	return configs, total, err
}

// GetByKey 根据 Key 获取配置
func (r *Repository) GetByKey(ctx context.Context, configKey string) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).Where("config_key = ? AND is_deleted = 0", configKey).First(&config).Error
	return &config, err
}

// Get 根据 ID 获取配置
func (r *Repository) Get(ctx context.Context, id int64) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&config).Error
	return &config, err
}

// Create 创建配置（ctx 携带操作人，由审计钩子填充 create_by）
func (r *Repository) Create(ctx context.Context, config *model.Config) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新配置
// 用 BuildPatchMap(form) 生成「列名→值」映射，从根上解决 GORM Updates(struct) 默认跳过零值的问题。
func (r *Repository) Update(ctx context.Context, form *model.ConfigForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Config{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// Delete 软删除配置
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Config{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// BatchDelete 批量软删除配置
func (r *Repository) BatchDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.Config{}).Where("id IN ?", ids).Update("is_deleted", 1).Error
}
