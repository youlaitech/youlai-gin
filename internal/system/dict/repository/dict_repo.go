package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/system/dict/model"
	"youlai-gin/pkg/gormx"
)

// Repository 字典数据访问层（含字典与字典项两个实体）
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Page 字典分页查询
func (r *Repository) Page(ctx context.Context, query *model.DictQuery) ([]model.Dict, int64, error) {
	var dicts []model.Dict
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Dict{}).Where("is_deleted = 0")
	if query.Keywords != "" {
		db = db.Where("dict_code LIKE ? OR name LIKE ?", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("create_time DESC").
		Find(&dicts).Error

	return dicts, total, err
}

// Options 启用中的字典列表（下拉选项来源）
func (r *Repository) Options(ctx context.Context) ([]model.Dict, error) {
	var dicts []model.Dict
	err := r.db.WithContext(ctx).Model(&model.Dict{}).
		Where("status = 1 AND is_deleted = 0").
		Order("create_time DESC").
		Find(&dicts).Error
	return dicts, err
}

// Get 根据ID查询字典
func (r *Repository) Get(ctx context.Context, id int64) (*model.Dict, error) {
	var dict model.Dict
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&dict).Error
	return &dict, err
}

// Create 创建字典（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) Create(ctx context.Context, dict *model.Dict) error {
	return r.db.WithContext(ctx).Create(dict).Error
}

// Update 更新字典
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入，
// 从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) Update(ctx context.Context, form *model.DictForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Dict{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// Delete 删除字典（逻辑删除）
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Dict{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CodeExists 检查字典编码是否存在（excludeId>0 时排除自身）
func (r *Repository) CodeExists(ctx context.Context, dictCode string, excludeId int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Dict{}).Where("dict_code = ? AND is_deleted = 0", dictCode)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// Items 根据字典编码获取字典项列表
func (r *Repository) Items(ctx context.Context, dictCode string) ([]model.DictItem, error) {
	var items []model.DictItem
	err := r.db.WithContext(ctx).Model(&model.DictItem{}).
		Where("dict_code = ? AND is_deleted = 0", dictCode).
		Order("sort ASC, id ASC").
		Find(&items).Error
	return items, err
}

// ItemPage 字典项分页查询
func (r *Repository) ItemPage(ctx context.Context, query *model.DictItemQuery) ([]model.DictItem, int64, error) {
	var items []model.DictItem
	var total int64

	db := r.db.WithContext(ctx).Model(&model.DictItem{}).Where("is_deleted = 0")
	if query.DictCode != "" {
		db = db.Where("dict_code = ?", query.DictCode)
	}
	if query.Keywords != "" {
		kw := "%" + query.Keywords + "%"
		db = db.Where("label LIKE ? OR value LIKE ?", kw, kw)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("sort ASC, id ASC").
		Find(&items).Error

	return items, total, err
}

// GetItem 根据ID查询字典项
func (r *Repository) GetItem(ctx context.Context, id int64) (*model.DictItem, error) {
	var item model.DictItem
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&item).Error
	return &item, err
}

// CreateItem 创建字典项（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) CreateItem(ctx context.Context, item *model.DictItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// UpdateItem 更新字典项
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入。
func (r *Repository) UpdateItem(ctx context.Context, form *model.DictItemForm) error {
	return r.db.WithContext(ctx).
		Model(&model.DictItem{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// DeleteItem 删除字典项（逻辑删除，保留历史数据）
func (r *Repository) DeleteItem(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.DictItem{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("is_deleted", 1).Error
}

// BatchDeleteItems 批量删除字典项（逻辑删除，保留历史数据）
func (r *Repository) BatchDeleteItems(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.DictItem{}).
		Where("id IN ?", ids).
		Update("is_deleted", 1).Error
}

// BatchDeleteItemsByCode 按字典编码级联逻辑删除字典项（删除字典时调用）
func (r *Repository) BatchDeleteItemsByCode(ctx context.Context, dictCode string) error {
	return r.db.WithContext(ctx).Model(&model.DictItem{}).
		Where("dict_code = ? AND is_deleted = 0", dictCode).
		Update("is_deleted", 1).Error
}

// ItemValueExists 同字典下值是否已存在（excludeID > 0 时排除自身，用于唯一性校验）
func (r *Repository) ItemValueExists(ctx context.Context, dictCode, value string, excludeID int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.DictItem{}).
		Where("dict_code = ? AND value = ? AND is_deleted = 0", dictCode, value)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ItemsCount 获取字典项数量（用于删除前校验）
func (r *Repository) ItemsCount(ctx context.Context, dictCode string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DictItem{}).Where("dict_code = ? AND is_deleted = 0", dictCode).Count(&count).Error
	return count, err
}
