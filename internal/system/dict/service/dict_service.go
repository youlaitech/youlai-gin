package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	sse "youlai-gin/internal/message/service"
	"youlai-gin/internal/system/dict/model"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Repository 字典数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.DictQuery) ([]model.Dict, int64, error)
	Options(ctx context.Context) ([]model.Dict, error)
	Get(ctx context.Context, id int64) (*model.Dict, error)
	Create(ctx context.Context, dict *model.Dict) error
	Update(ctx context.Context, form *model.DictForm) error
	Delete(ctx context.Context, id int64) error
	CodeExists(ctx context.Context, dictCode string, excludeId int64) (bool, error)
	Items(ctx context.Context, dictCode string) ([]model.DictItem, error)
	ItemPage(ctx context.Context, query *model.DictItemQuery) ([]model.DictItem, int64, error)
	GetItem(ctx context.Context, id int64) (*model.DictItem, error)
	CreateItem(ctx context.Context, item *model.DictItem) error
	UpdateItem(ctx context.Context, form *model.DictItemForm) error
	DeleteItem(ctx context.Context, id int64) error
	BatchDeleteItems(ctx context.Context, ids []int64) error
	ItemsCount(ctx context.Context, dictCode string) (int64, error)
}

// Service 字典业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Page 字典分页列表
func (s *Service) Page(ctx context.Context, query *model.DictQuery) (*baseModel.PagedData, error) {
	dicts, total, err := s.repo.Page(ctx, query)
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

	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// Options 字典下拉选项
func (s *Service) Options(ctx context.Context) ([]baseModel.Option[string], error) {
	dicts, err := s.repo.Options(ctx)
	if err != nil {
		return nil, errs.SystemError("查询字典列表失败")
	}

	options := make([]baseModel.Option[string], len(dicts))
	for i, dict := range dicts {
		options[i] = baseModel.Option[string]{Value: dict.DictCode, Label: dict.Name}
	}

	return options, nil
}

// Create 新增字典
func (s *Service) Create(ctx context.Context, form *model.DictForm) error {
	if err := s.checkCodeUnique(ctx, form.DictCode, 0); err != nil {
		return err
	}

	dict := &model.Dict{
		ID:       form.ID,
		DictCode: form.DictCode,
		Name:     form.Name,
		Status:   form.Status,
		Remark:   form.Remark,
	}
	if err := s.repo.Create(ctx, dict); err != nil {
		return errs.SystemError("创建字典失败")
	}
	form.ID = dict.ID

	s.notifyDictChange(form.DictCode)
	return nil
}

// Update 更新字典
func (s *Service) Update(ctx context.Context, id int64, form *model.DictForm) error {
	form.ID = types.BigInt(id)

	if err := s.checkCodeUnique(ctx, form.DictCode, id); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新字典失败")
	}

	s.notifyDictChange(form.DictCode)
	return nil
}

// checkCodeUnique 字典编码唯一性校验（excludeId 排除自身）
func (s *Service) checkCodeUnique(ctx context.Context, dictCode string, excludeId int64) error {
	exists, err := s.repo.CodeExists(ctx, dictCode, excludeId)
	if err != nil {
		return errs.SystemError("检查字典编码失败")
	}
	if exists {
		return errs.Business("字典编码已存在")
	}
	return nil
}

// GetForm 获取字典表单数据
func (s *Service) GetForm(ctx context.Context, id int64) (*model.DictForm, error) {
	dict, err := s.repo.Get(ctx, id)
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

// Delete 删除字典（存在字典项时拒绝）
func (s *Service) Delete(ctx context.Context, id int64) error {
	dict, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("字典不存在")
		}
		return errs.SystemError("查询字典失败")
	}

	count, err := s.repo.ItemsCount(ctx, dict.DictCode)
	if err != nil {
		return errs.SystemError("查询字典项失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除该字典下的所有字典项")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errs.SystemError("删除字典失败")
	}

	s.notifyDictChange(dict.DictCode)
	return nil
}

// Items 获取字典项列表
func (s *Service) Items(ctx context.Context, dictCode string) ([]model.DictItemVO, error) {
	items, err := s.repo.Items(ctx, dictCode)
	if err != nil {
		return nil, errs.SystemError("查询字典项失败")
	}

	voList := make([]model.DictItemVO, len(items))
	for i, item := range items {
		voList[i] = toItemVO(item)
	}

	return voList, nil
}

// ItemPage 字典项分页列表
func (s *Service) ItemPage(ctx context.Context, query *model.DictItemQuery) (*baseModel.PagedData, error) {
	items, total, err := s.repo.ItemPage(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询字典项列表失败")
	}

	voList := make([]model.DictItemVO, len(items))
	for i, item := range items {
		voList[i] = toItemVO(item)
	}

	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// CreateItem 新增字典项
func (s *Service) CreateItem(ctx context.Context, form *model.DictItemForm) error {
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
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return errs.SystemError("创建字典项失败")
	}
	form.ID = item.ID

	s.notifyDictChange(form.DictCode)
	return nil
}

// UpdateItem 更新字典项
func (s *Service) UpdateItem(ctx context.Context, id int64, form *model.DictItemForm) error {
	form.ID = types.BigInt(id)

	if err := s.repo.UpdateItem(ctx, form); err != nil {
		return errs.SystemError("更新字典项失败")
	}

	s.notifyDictChange(form.DictCode)
	return nil
}

// GetItemForm 获取字典项表单数据
func (s *Service) GetItemForm(ctx context.Context, id int64) (*model.DictItemForm, error) {
	item, err := s.repo.GetItem(ctx, id)
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

// DeleteItem 删除字典项
func (s *Service) DeleteItem(ctx context.Context, id int64) error {
	item, err := s.repo.GetItem(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("字典项不存在")
		}
		return errs.SystemError("查询字典项失败")
	}

	if err := s.repo.DeleteItem(ctx, id); err != nil {
		return errs.SystemError("删除字典项失败")
	}

	s.notifyDictChange(item.DictCode)
	return nil
}

// BatchDeleteItems 批量删除字典项
func (s *Service) BatchDeleteItems(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errs.BadRequest("无效的字典项ID")
	}

	if err := s.repo.BatchDeleteItems(ctx, ids); err != nil {
		return errs.SystemError("删除字典项失败")
	}

	return nil
}

// toItemVO 字典项实体转视图
func toItemVO(item model.DictItem) model.DictItemVO {
	return model.DictItemVO{
		ID:      types.BigInt(item.ID),
		Value:   item.Value,
		Label:   item.Label,
		TagType: item.TagType,
		Sort:    item.Sort,
		Status:  item.Status,
		Remark:  item.Remark,
	}
}

// notifyDictChange 推送字典变更事件（SSE 未就绪时静默跳过）
func (s *Service) notifyDictChange(dictCode string) {
	if sse := sse.GetSseService(); sse != nil {
		sse.SendDictChange(dictCode)
	}
}
