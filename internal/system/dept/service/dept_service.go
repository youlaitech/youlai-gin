package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/utils"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// Repository 部门数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	List(ctx context.Context, query *model.DeptQuery, currentUser *auth.UserDetails) ([]model.Dept, error)
	Get(ctx context.Context, id int64) (*model.Dept, error)
	Create(ctx context.Context, dept *model.Dept) error
	Update(ctx context.Context, form *model.DeptForm) error
	Delete(ctx context.Context, id int64) error
	Options(ctx context.Context, currentUser *auth.UserDetails) ([]model.Dept, error)
	NameExists(ctx context.Context, name string, parentId int64, excludeId int64) (bool, error)
	CodeExists(ctx context.Context, code string, excludeId int64) (bool, error)
	ChildrenCount(ctx context.Context, parentId int64) (int64, error)
}

// Service 部门业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List 部门列表（树形结构）
func (s *Service) List(ctx context.Context, query *model.DeptQuery, currentUser *auth.UserDetails) ([]*model.DeptVO, error) {
	depts, err := s.repo.List(ctx, query, currentUser)
	if err != nil {
		return nil, errs.SystemError("查询部门列表失败")
	}

	deptVOs := make([]*model.DeptVO, len(depts))
	for i, dept := range depts {
		deptVOs[i] = &model.DeptVO{
			ID:         dept.ID,
			Name:       dept.Name,
			Code:       dept.Code,
			ParentID:   dept.ParentID,
			TreePath:   dept.TreePath,
			Sort:       dept.Sort,
			Status:     dept.Status,
			CreateTime: types.LocalTime(dept.CreateTime),
			UpdateTime: types.LocalTime(dept.UpdateTime),
		}
	}

	tree := utils.BuildTreeSimple(
		deptVOs,
		func(d *model.DeptVO) int64 { return int64(d.ID) },
		func(d *model.DeptVO) int64 { return int64(d.ParentID) },
		func(d **model.DeptVO, children []*model.DeptVO) {
			(*d).Children = children
		},
	)

	return tree, nil
}

// Options 部门下拉选项（树形结构）
func (s *Service) Options(ctx context.Context, currentUser *auth.UserDetails) ([]model.DeptOption, error) {
	depts, err := s.repo.Options(ctx, currentUser)
	if err != nil {
		return nil, errs.SystemError("查询部门下拉失败")
	}

	options := make([]model.DeptOption, len(depts))
	for i, dept := range depts {
		options[i] = model.DeptOption{
			Value:    types.BigInt(dept.ID),
			Label:    dept.Name,
			ParentID: types.BigInt(dept.ParentID),
		}
	}

	tree := utils.BuildTreeSimple(
		options,
		func(d model.DeptOption) int64 { return int64(d.Value) },
		func(d model.DeptOption) int64 { return int64(d.ParentID) },
		func(d *model.DeptOption, children []model.DeptOption) {
			childNodes := make([]*model.DeptOption, len(children))
			for i := range children {
				child := children[i]
				childNodes[i] = &child
			}
			d.Children = childNodes
		},
	)

	return tree, nil
}

// Create 新增部门（按父部门拼接 tree_path）
func (s *Service) Create(ctx context.Context, form *model.DeptForm) error {
	if err := s.checkUnique(ctx, form, 0); err != nil {
		return err
	}

	var treePath string
	if form.ParentID == 0 {
		treePath = "0"
	} else {
		parent, err := s.repo.Get(ctx, int64(form.ParentID))
		if err != nil {
			return errs.SystemError("查询父部门失败")
		}
		treePath = fmt.Sprintf("%s,%d", parent.TreePath, parent.ID)
	}

	dept := &model.Dept{
		ID:       form.ID,
		Name:     form.Name,
		Code:     form.Code,
		ParentID: form.ParentID,
		TreePath: treePath,
		Sort:     form.Sort,
		Status:   form.Status,
	}
	if err := s.repo.Create(ctx, dept); err != nil {
		return errs.SystemError("创建部门失败")
	}
	form.ID = dept.ID

	return nil
}

// Update 更新部门
func (s *Service) Update(ctx context.Context, id int64, form *model.DeptForm) error {
	form.ID = types.BigInt(id)

	if err := s.checkUnique(ctx, form, id); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新部门失败")
	}

	return nil
}

// checkUnique 同级名称与编码唯一性校验（excludeId 排除自身）
func (s *Service) checkUnique(ctx context.Context, form *model.DeptForm, excludeId int64) error {
	exists, err := s.repo.NameExists(ctx, form.Name, int64(form.ParentID), excludeId)
	if err != nil {
		return errs.SystemError("检查部门名称失败")
	}
	if exists {
		return errs.BadRequest("同级部门名称已存在")
	}

	exists, err = s.repo.CodeExists(ctx, form.Code, excludeId)
	if err != nil {
		return errs.SystemError("检查部门编号失败")
	}
	if exists {
		return errs.Business("部门编号已存在")
	}

	return nil
}

// GetForm 获取部门表单数据
func (s *Service) GetForm(ctx context.Context, id int64) (*model.DeptForm, error) {
	dept, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("部门不存在")
		}
		return nil, errs.SystemError("查询部门失败")
	}

	return &model.DeptForm{
		ID:       dept.ID,
		Name:     dept.Name,
		Code:     dept.Code,
		ParentID: dept.ParentID,
		Sort:     dept.Sort,
		Status:   dept.Status,
	}, nil
}

// Delete 删除部门（存在子部门时拒绝）
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.Get(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("部门不存在")
		}
		return errs.SystemError("查询部门失败")
	}

	count, err := s.repo.ChildrenCount(ctx, id)
	if err != nil {
		return errs.SystemError("查询子部门失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除子部门")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errs.SystemError("删除部门失败")
	}

	return nil
}
