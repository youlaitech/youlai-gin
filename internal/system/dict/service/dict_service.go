package service

import (
	"errors"

	"gorm.io/gorm"

	"youlai-gin/internal/system/dict/model"
	"youlai-gin/internal/system/dict/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// GetDictPage 字典分页列表
func GetDictPage(query *model.DictQuery) (*common.PagedData, error) {
	dicts, total, err := repository.GetDictPage(query)
	if err != nil {
		return nil, errs.SystemError("查询字典列表失败")
	}

	voList := make([]model.DictPageVO, len(dicts))
	for i, dict := range dicts {
		voList[i] = model.DictPageVO{
			ID:         types.BigInt(dict.ID),
			DictCode:   dict.DictCode,
			Name:       dict.Name,
			Status:     dict.Status,
			Remark:     dict.Remark,
			CreateTime: types.LocalTime(dict.CreateTime),
			UpdateTime: types.LocalTime(dict.UpdateTime),
		}
	}

	return &common.PagedData{List: voList, Total: total}, nil
}

// GetDictList 获取字典下拉选项
func GetDictList() ([]common.Option[string], error) {
	dicts, err := repository.GetDictList()
	if err != nil {
		return nil, errs.SystemError("查询字典列表失败")
	}

	options := make([]common.Option[string], len(dicts))
	for i, dict := range dicts {
		options[i] = common.Option[string]{
			Value: dict.DictCode,
			Label: dict.Name,
		}
	}

	return options, nil
}

// SaveDict 保存字典（新增或更新）
func SaveDict(form *model.DictForm) error {
	exists, err := repository.CheckDictCodeExists(form.DictCode, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查字典编码失败")
	}
	if exists {
		return errs.BadRequest("字典编码已存在")
	}

	dict := &model.Dict{
		ID:       form.ID,
		DictCode: form.DictCode,
		Name:     form.Name,
		Status:   form.Status,
		Remark:   form.Remark,
	}

	if form.ID == 0 {
		if err := repository.CreateDict(dict); err != nil {
			return errs.SystemError("创建字典失败")
		}
	} else {
		if err := repository.UpdateDict(dict); err != nil {
			return errs.SystemError("更新字典失败")
		}
	}

	return nil
}

// BatchDeleteDictItems 批量删除字典项
func BatchDeleteDictItems(ids []int64) error {
	if len(ids) == 0 {
		return errs.BadRequest("无效的字典项ID")
	}

	if err := repository.BatchDeleteDictItems(ids); err != nil {
		return errs.SystemError("删除字典项失败")
	}

	return nil
}

// GetDictForm 获取字典表单数据
func GetDictForm(id int64) (*model.DictForm, error) {
	dict, err := repository.GetDictByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("字典不存在")
		}
		return nil, errs.SystemError("查询字典失败")
	}

	return &model.DictForm{
		ID:       dict.ID,
		DictCode: dict.DictCode,
		Name:     dict.Name,
		Status:   dict.Status,
		Remark:   dict.Remark,
	}, nil
}

// DeleteDict 删除字典
func DeleteDict(id int64) error {
	dict, err := repository.GetDictByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("字典不存在")
		}
		return errs.SystemError("查询字典失败")
	}

	count, err := repository.GetDictItemsCount(dict.DictCode)
	if err != nil {
		return errs.SystemError("查询字典项失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除该字典下的所有字典项")
	}

	if err := repository.DeleteDict(id); err != nil {
		return errs.SystemError("删除字典失败")
	}

	return nil
}

// GetDictItems 获取字典项列表
func GetDictItems(dictCode string) ([]model.DictItemVO, error) {
	items, err := repository.GetDictItems(dictCode)
	if err != nil {
		return nil, errs.SystemError("查询字典项失败")
	}

	voList := make([]model.DictItemVO, len(items))
	for i, item := range items {
		voList[i] = model.DictItemVO{
			ID:     types.BigInt(item.ID),
			Value:  item.Value,
			Label:  item.Label,
			TagType: item.TagType,
			Sort:   item.Sort,
			Status: item.Status,
			Remark: item.Remark,
		}
	}

	return voList, nil
}

// GetDictItemPage 字典项分页列表
func GetDictItemPage(query *model.DictItemQuery) (*common.PagedData, error) {
	items, total, err := repository.GetDictItemPage(query)
	if err != nil {
		return nil, errs.SystemError("查询字典项列表失败")
	}

	voList := make([]model.DictItemVO, len(items))
	for i, item := range items {
		voList[i] = model.DictItemVO{
			ID:     types.BigInt(item.ID),
			Value:  item.Value,
			Label:  item.Label,
			TagType: item.TagType,
			Sort:   item.Sort,
			Status: item.Status,
			Remark: item.Remark,
		}
	}

	return &common.PagedData{List: voList, Total: total}, nil
}

// SaveDictItem 保存字典项（新增或更新）
func SaveDictItem(form *model.DictItemForm) error {
	item := &model.DictItem{
		ID:       form.ID,
		DictCode: form.DictCode,
		Value:    form.Value,
		Label:    form.Label,
		TagType:  form.TagType,
		Sort:     form.Sort,
		Status:   form.Status,
		Remark:   form.Remark,
	}

	if form.ID == 0 {
		if err := repository.CreateDictItem(item); err != nil {
			return errs.SystemError("创建字典项失败")
		}
	} else {
		if err := repository.UpdateDictItem(item); err != nil {
			return errs.SystemError("更新字典项失败")
		}
	}

	return nil
}

// GetDictItemForm 获取字典项表单数据
func GetDictItemForm(id int64) (*model.DictItemForm, error) {
	item, err := repository.GetDictItemByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("字典项不存在")
		}
		return nil, errs.SystemError("查询字典项失败")
	}

	return &model.DictItemForm{
		ID:       item.ID,
		DictCode: item.DictCode,
		Value:    item.Value,
		Label:    item.Label,
		TagType:  item.TagType,
		Sort:     item.Sort,
		Status:   item.Status,
		Remark:   item.Remark,
	}, nil
}

// DeleteDictItem 删除字典项
func DeleteDictItem(id int64) error {
	_, err := repository.GetDictItemByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("字典项不存在")
		}
		return errs.SystemError("查询字典项失败")
	}

	if err := repository.DeleteDictItem(id); err != nil {
		return errs.SystemError("删除字典项失败")
	}

	return nil
}
