package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/system/dict/model"
	"youlai-gin/pkg/gormx"
)

// Repository 字典数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// GetDictPage 字典分页查询
func (r *Repository) GetDictPage(query *model.DictQuery) ([]model.Dict, int64, error) {
	var dicts []model.Dict
	var total int64
	db := r.db.Model(&model.Dict{}).Where("is_deleted = 0")
	if query.Keywords != "" {
		db = db.Where("dict_code LIKE ? OR name LIKE ?", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := query.GetOffset()
	limit := query.GetLimit()
	if err := db.Offset(offset).Limit(limit).Order("create_time DESC").Find(&dicts).Error; err != nil {
		return nil, 0, err
	}
	return dicts, total, nil
}

// GetDictList 获取字典列表
func (r *Repository) GetDictList() ([]model.Dict, error) {
	var dicts []model.Dict
	err := r.db.Model(&model.Dict{}).Where("status = 1 AND is_deleted = 0").Order("create_time DESC").Find(&dicts).Error
	return dicts, err
}

// GetDictByID 根据ID查询字典
func (r *Repository) GetDictByID(id int64) (*model.Dict, error) {
	var dict model.Dict
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&dict).Error
	return &dict, err
}

// CreateDict 创建字典（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) CreateDict(ctx context.Context, dict *model.Dict) error {
	return r.db.WithContext(ctx).Create(dict).Error
}

// UpdateDict 更新字典
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入，
// 从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) UpdateDict(ctx context.Context, form *model.DictForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Dict{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// DeleteDict 删除字典（逻辑删除）
func (r *Repository) DeleteDict(id int64) error {
	return r.db.Model(&model.Dict{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CheckDictCodeExists 检查字典编码是否存在
func (r *Repository) CheckDictCodeExists(dictCode string, excludeId int64) (bool, error) {
	var count int64
	db := r.db.Model(&model.Dict{}).Where("dict_code = ? AND is_deleted = 0", dictCode)
	if excludeId > 0 { db = db.Where("id != ?", excludeId) }
	err := db.Count(&count).Error
	return count > 0, err
}

// GetDictItems 根据字典编码获取字典项列表
func (r *Repository) GetDictItems(dictCode string) ([]model.DictItem, error) {
	var items []model.DictItem
	err := r.db.Model(&model.DictItem{}).Where("dict_code = ?", dictCode).Order("sort ASC, id ASC").Find(&items).Error
	return items, err
}

// GetDictItemPage 字典项分页查询
func (r *Repository) GetDictItemPage(query *model.DictItemQuery) ([]model.DictItem, int64, error) {
	var items []model.DictItem
	var total int64
	db := r.db.Model(&model.DictItem{})
	if query.DictCode != "" { db = db.Where("dict_code = ?", query.DictCode) }
	if query.Keywords != "" {
		kw := "%" + query.Keywords + "%"
		db = db.Where("label LIKE ? OR value LIKE ?", kw, kw)
	}
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	offset := query.GetOffset()
	limit := query.GetLimit()
	if err := db.Offset(offset).Limit(limit).Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetDictItemByID 根据ID查询字典项
func (r *Repository) GetDictItemByID(id int64) (*model.DictItem, error) {
	var item model.DictItem
	err := r.db.Where("id = ?", id).First(&item).Error
	return &item, err
}

// CreateDictItem 创建字典项（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) CreateDictItem(ctx context.Context, item *model.DictItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// UpdateDictItem 更新字典项
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入。
func (r *Repository) UpdateDictItem(ctx context.Context, form *model.DictItemForm) error {
	return r.db.WithContext(ctx).
		Model(&model.DictItem{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// DeleteDictItem 删除字典项（物理删除）
func (r *Repository) DeleteDictItem(id int64) error { return r.db.Where("id = ?", id).Delete(&model.DictItem{}).Error }

// BatchDeleteDictItems 批量删除字典项
func (r *Repository) BatchDeleteDictItems(ids []int64) error { return r.db.Where("id IN ?", ids).Delete(&model.DictItem{}).Error }

// GetDictItemsCount 获取字典项数量（用于删除前校验）
func (r *Repository) GetDictItemsCount(dictCode string) (int64, error) {
	var count int64
	err := r.db.Model(&model.DictItem{}).Where("dict_code = ?", dictCode).Count(&count).Error
	return count, err
}
