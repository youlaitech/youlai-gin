package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/system/dept/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/utils"
)

// GetDeptList 部门列表（树形结构）
func GetDeptList(query *model.DeptQuery) ([]*model.DeptVO, error) {
	depts, err := repository.GetDeptList(query)
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
			CreateTime: dept.CreateTime,
			UpdateTime: dept.UpdateTime,
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
func GetDeptOptions() ([]common.Option[types.BigInt], error) {
	depts, err := repository.GetDeptOptions()
	if err != nil {
		return nil, errs.SystemError("查询部门选项失败")
	}

	options := make([]common.Option[types.BigInt], len(depts))
	for i, dept := range depts {
		options[i] = common.Option[types.BigInt]{
			Value: types.BigInt(dept.ID),
			Label: dept.Name,
		}
	}

	return options, nil
}

// SaveDept 保存部门（新增或更新）
func SaveDept(form *model.DeptForm) error {
	exists, err := repository.CheckDeptNameExists(form.Name, int64(form.ParentID), int64(form.ID))
	if err != nil {
		return errs.SystemError("检查部门名称失败")
	}
	if exists {
		return errs.BadRequest("同级部门名称已存在")
	}

	exists, err = repository.CheckDeptCodeExists(form.Code, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查部门编号失败")
	}
	if exists {
		return errs.BadRequest("部门编号已存在")
	}

	dept := &model.Dept{
		ID:       form.ID,
		Name:     form.Name,
		Code:     form.Code,
		ParentID: form.ParentID,
		Sort:     form.Sort,
		Status:   form.Status,
	}

	if form.ParentID == 0 {
		dept.TreePath = "0"
	} else {
		parent, err := repository.GetDeptByID(int64(form.ParentID))
		if err != nil {
			return errs.SystemError("查询父部门失败")
		}
		dept.TreePath = fmt.Sprintf("%s,%d", parent.TreePath, parent.ID)
	}

	if form.ID == 0 {
		if err := repository.CreateDept(dept); err != nil {
			return errs.SystemError("创建部门失败")
		}
	} else {
		if err := repository.UpdateDept(dept); err != nil {
			return errs.SystemError("更新部门失败")
		}
	}

	return nil
}

// GetDeptForm 获取部门表单数据
func GetDeptForm(id int64) (*model.DeptForm, error) {
	dept, err := repository.GetDeptByID(id)
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
func DeleteDept(id int64) error {
	_, err := repository.GetDeptByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("部门不存在")
		}
		return errs.SystemError("查询部门失败")
	}

	count, err := repository.GetChildrenCount(id)
	if err != nil {
		return errs.SystemError("查询子部门失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除子部门")
	}

	if err := repository.DeleteDept(id); err != nil {
		return errs.SystemError("删除部门失败")
	}

	return nil
}
