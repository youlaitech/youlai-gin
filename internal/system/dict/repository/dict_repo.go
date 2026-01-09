package repository

import (
	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/dict/model"
)

// GetDictPage 字典分页查询
func GetDictPage(query *model.DictQuery) ([]model.Dict, int64, error) {
	var dicts []model.Dict
	var total int64

	db := database.DB.Model(&model.Dict{}).Where("is_deleted = 0")

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
func GetDictList() ([]model.Dict, error) {
	var dicts []model.Dict
	err := database.DB.Model(&model.Dict{}).
		Where("status = 1 AND is_deleted = 0").
		Order("create_time DESC").
		Find(&dicts).Error
	return dicts, err
}

// GetDictByID 根据ID查询字典
func GetDictByID(id int64) (*model.Dict, error) {
	var dict model.Dict
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&dict).Error
	return &dict, err
}

// CreateDict 创建字典
func CreateDict(dict *model.Dict) error {
	return database.DB.Create(dict).Error
}

// UpdateDict 更新字典
func UpdateDict(dict *model.Dict) error {
	return database.DB.Model(&model.Dict{}).Where("id = ?", dict.ID).Updates(dict).Error
}

// DeleteDict 删除字典（逻辑删除）
func DeleteDict(id int64) error {
	return database.DB.Model(&model.Dict{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CheckDictCodeExists 检查字典编码是否存在
func CheckDictCodeExists(dictCode string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Dict{}).Where("dict_code = ? AND is_deleted = 0", dictCode)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetDictItems 根据字典编码获取字典项列表
func GetDictItems(dictCode string) ([]model.DictItem, error) {
	var items []model.DictItem
	err := database.DB.Model(&model.DictItem{}).
		Where("dict_code = ?", dictCode).
		Order("sort ASC, id ASC").
		Find(&items).Error
	return items, err
}

// GetDictItemPage 字典项分页查询
func GetDictItemPage(query *model.DictItemQuery) ([]model.DictItem, int64, error) {
	var items []model.DictItem
	var total int64

	db := database.DB.Model(&model.DictItem{})
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

	offset := query.GetOffset()
	limit := query.GetLimit()
	if err := db.Offset(offset).Limit(limit).Order("sort ASC, id ASC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetDictItemByID 根据ID查询字典项
func GetDictItemByID(id int64) (*model.DictItem, error) {
	var item model.DictItem
	err := database.DB.Where("id = ?", id).First(&item).Error
	return &item, err
}

// CreateDictItem 创建字典项
func CreateDictItem(item *model.DictItem) error {
	return database.DB.Create(item).Error
}

// UpdateDictItem 更新字典项
func UpdateDictItem(item *model.DictItem) error {
	return database.DB.Model(&model.DictItem{}).Where("id = ?", item.ID).Select(
		"dict_code",
		"value",
		"label",
		"tag_type",
		"sort",
		"status",
		"remark",
	).Updates(item).Error
}

// DeleteDictItem 删除字典项（物理删除）
func DeleteDictItem(id int64) error {
	return database.DB.Where("id = ?", id).Delete(&model.DictItem{}).Error
}

// BatchDeleteDictItems 批量删除字典项
func BatchDeleteDictItems(ids []int64) error {
	return database.DB.Where("id IN ?", ids).Delete(&model.DictItem{}).Error
}

// GetDictItemsCount 获取字典项数量（用于删除前校验）
func GetDictItemsCount(dictCode string) (int64, error) {
	var count int64
	err := database.DB.Model(&model.DictItem{}).
		Where("dict_code = ?", dictCode).
		Count(&count).Error
	return count, err
}
