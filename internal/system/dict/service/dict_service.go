package service

import (
	"errors"

	"gorm.io/gorm"

	"youlai-gin/internal/message"
	"youlai-gin/internal/system/dict/model"
	"youlai-gin/internal/system/dict/repository"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// Service 字典业务逻辑层
type Service struct {
	repo *repository.Repository
}

// NewService 创建 Service 实例
func NewService(repo *repository.Repository) *Service { return &Service{repo: repo} }

// GetDictPage 字典分页列表
func (s *Service) GetDictPage(query *model.DictQuery) (*baseModel.PagedData, error) {
	dicts, total, err := s.repo.GetDictPage(query)
	if err != nil { return nil, errs.SystemError("查询字典列表失败") }
	voList := make([]model.DictPageVO, len(dicts))
	for i, dict := range dicts {
		voList[i] = model.DictPageVO{ID: types.BigInt(dict.ID), DictCode: dict.DictCode, Name: dict.Name, Status: dict.Status, Remark: dict.Remark, CreateTime: types.LocalTime(dict.CreateTime), UpdateTime: types.LocalTime(dict.UpdateTime)}
	}
	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// GetDictList 获取字典下拉选项
func (s *Service) GetDictList() ([]baseModel.Option[string], error) {
	dicts, err := s.repo.GetDictList()
	if err != nil { return nil, errs.SystemError("查询字典列表失败") }
	options := make([]baseModel.Option[string], len(dicts))
	for i, dict := range dicts { options[i] = baseModel.Option[string]{Value: dict.DictCode, Label: dict.Name} }
	return options, nil
}

// SaveDict 保存字典（新增或更新）
func (s *Service) SaveDict(form *model.DictForm) error {
	exists, err := s.repo.CheckDictCodeExists(form.DictCode, int64(form.ID))
	if err != nil { return errs.SystemError("检查字典编码失败") }
	if exists { return errs.Business("字典编码已存在") }
	dict := &model.Dict{ID: form.ID, DictCode: form.DictCode, Name: form.Name, Status: form.Status, Remark: form.Remark}
	if form.ID == 0 {
		if err := s.repo.CreateDict(dict); err != nil { return errs.SystemError("创建字典失败") }
	} else {
		if err := s.repo.UpdateDict(dict); err != nil { return errs.SystemError("更新字典失败") }
	}
	if sse := message.GetSseService(); sse != nil { sse.SendDictChange(form.DictCode) }
	return nil
}

// BatchDeleteDictItems 批量删除字典项
func (s *Service) BatchDeleteDictItems(ids []int64) error {
	if len(ids) == 0 { return errs.BadRequest("无效的字典项ID") }
	if err := s.repo.BatchDeleteDictItems(ids); err != nil { return errs.SystemError("删除字典项失败") }
	return nil
}

// GetDictForm 获取字典表单数据
func (s *Service) GetDictForm(id int64) (*model.DictForm, error) {
	dict, err := s.repo.GetDictByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, errs.NotFound("字典不存在") }
		return nil, errs.SystemError("查询字典失败")
	}
	return &model.DictForm{ID: dict.ID, DictCode: dict.DictCode, Name: dict.Name, Status: dict.Status, Remark: dict.Remark}, nil
}

// DeleteDict 删除字典
func (s *Service) DeleteDict(id int64) error {
	dict, err := s.repo.GetDictByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errs.NotFound("字典不存在") }
		return errs.SystemError("查询字典失败")
	}
	count, err := s.repo.GetDictItemsCount(dict.DictCode)
	if err != nil { return errs.SystemError("查询字典项失败") }
	if count > 0 { return errs.BadRequest("请先删除该字典下的所有字典项") }
	if err := s.repo.DeleteDict(id); err != nil { return errs.SystemError("删除字典失败") }
	if sse := message.GetSseService(); sse != nil { sse.SendDictChange(dict.DictCode) }
	return nil
}

// GetDictItems 获取字典项列表
func (s *Service) GetDictItems(dictCode string) ([]model.DictItemVO, error) {
	items, err := s.repo.GetDictItems(dictCode)
	if err != nil { return nil, errs.SystemError("查询字典项失败") }
	voList := make([]model.DictItemVO, len(items))
	for i, item := range items { voList[i] = model.DictItemVO{ID: types.BigInt(item.ID), Value: item.Value, Label: item.Label, TagType: item.TagType, Sort: item.Sort, Status: item.Status, Remark: item.Remark} }
	return voList, nil
}

// GetDictItemPage 字典项分页列表
func (s *Service) GetDictItemPage(query *model.DictItemQuery) (*baseModel.PagedData, error) {
	items, total, err := s.repo.GetDictItemPage(query)
	if err != nil { return nil, errs.SystemError("查询字典项列表失败") }
	voList := make([]model.DictItemVO, len(items))
	for i, item := range items { voList[i] = model.DictItemVO{ID: types.BigInt(item.ID), Value: item.Value, Label: item.Label, TagType: item.TagType, Sort: item.Sort, Status: item.Status, Remark: item.Remark} }
	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// SaveDictItem 保存字典项（新增或更新）
func (s *Service) SaveDictItem(form *model.DictItemForm) error {
	item := &model.DictItem{ID: form.ID, DictCode: form.DictCode, Value: form.Value, Label: form.Label, TagType: form.TagType, Sort: form.Sort, Status: form.Status, Remark: form.Remark}
	if form.ID == 0 {
		if err := s.repo.CreateDictItem(item); err != nil { return errs.SystemError("创建字典项失败") }
	} else {
		if err := s.repo.UpdateDictItem(item); err != nil { return errs.SystemError("更新字典项失败") }
	}
	if sse := message.GetSseService(); sse != nil { sse.SendDictChange(form.DictCode) }
	return nil
}

// GetDictItemForm 获取字典项表单数据
func (s *Service) GetDictItemForm(id int64) (*model.DictItemForm, error) {
	item, err := s.repo.GetDictItemByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, errs.NotFound("字典项不存在") }
		return nil, errs.SystemError("查询字典项失败")
	}
	return &model.DictItemForm{ID: item.ID, DictCode: item.DictCode, Value: item.Value, Label: item.Label, TagType: item.TagType, Sort: item.Sort, Status: item.Status, Remark: item.Remark}, nil
}

// DeleteDictItem 删除字典项
func (s *Service) DeleteDictItem(id int64) error {
	item, err := s.repo.GetDictItemByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errs.NotFound("字典项不存在") }
		return errs.SystemError("查询字典项失败")
	}
	if err := s.repo.DeleteDictItem(id); err != nil { return errs.SystemError("删除字典项失败") }
	if sse := message.GetSseService(); sse != nil { sse.SendDictChange(item.DictCode) }
	return nil
}
