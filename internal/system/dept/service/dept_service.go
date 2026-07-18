package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/common/auth"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
	"youlai-gin/internal/common/utils"
)

// Repository 部门数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	GetDeptList(query *model.DeptQuery, currentUser *auth.UserDetails) ([]model.Dept, error)
	GetDeptByID(id int64) (*model.Dept, error)
	CreateDept(ctx context.Context, dept *model.Dept) error
	UpdateDept(ctx context.Context, form *model.DeptForm) error
	DeleteDept(id int64) error
	GetDeptOptions(currentUser *auth.UserDetails) ([]model.Dept, error)
	CheckDeptNameExists(name string, parentId int64, excludeId int64) (bool, error)
	CheckDeptCodeExists(code string, excludeId int64) (bool, error)
	GetChildrenCount(parentId int64) (int64, error)
}

// Service 部门业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetDeptList 部门列表（树形结构）
func (s *Service) GetDeptList(query *model.DeptQuery, currentUser *auth.UserDetails) ([]*model.DeptVO, error) {
	depts, err := s.repo.GetDeptList(query, currentUser)
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

// GetDeptOptions 部门下拉选项
func (s *Service) GetDeptOptions(currentUser *auth.UserDetails) ([]model.DeptOption, error) {
	depts, err := s.repo.GetDeptOptions(currentUser)
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

// SaveDept 新增或更新部门
func (s *Service) SaveDept(c *gin.Context, form *model.DeptForm) error {
	exists, err := s.repo.CheckDeptNameExists(form.Name, int64(form.ParentID), int64(form.ID))
	if err != nil {
		return errs.SystemError("检查部门名称失败")
	}
	if exists {
		return errs.BadRequest("同级部门名称已存在")
	}

	exists, err = s.repo.CheckDeptCodeExists(form.Code, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查部门编号失败")
	}
	if exists {
		return errs.Business("部门编号已存在")
	}

	ctx := appContext.OperatorCtx(c)

	if form.ID == 0 {
		var treePath string
		if form.ParentID == 0 {
			treePath = "0"
		} else {
			parent, err := s.repo.GetDeptByID(int64(form.ParentID))
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
		if err := s.repo.CreateDept(ctx, dept); err != nil {
			return errs.SystemError("创建部门失败")
		}
	} else {
		if err := s.repo.UpdateDept(ctx, form); err != nil {
			return errs.SystemError("更新部门失败")
		}
	}

	return nil
}

// GetDeptForm 获取部门表单数据
func (s *Service) GetDeptForm(id int64) (*model.DeptForm, error) {
	dept, err := s.repo.GetDeptByID(id)
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



// DeleteDept 删除部门
func (s *Service) DeleteDept(id int64) error {
	_, err := s.repo.GetDeptByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("部门不存在")
		}
		return errs.SystemError("查询部门失败")
	}

	count, err := s.repo.GetChildrenCount(id)
	if err != nil {
		return errs.SystemError("查询子部门失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除子部门")
	}

	if err := s.repo.DeleteDept(id); err != nil {
		return errs.SystemError("删除部门失败")
	}

	return nil
}
