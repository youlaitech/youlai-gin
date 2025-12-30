package service

import (
	"errors"

	"gorm.io/gorm"

	"youlai-gin/internal/system/role/model"
	"youlai-gin/internal/system/role/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// GetRolePage 角色分页列表
func GetRolePage(query *model.RolePageQuery) (*common.PageResult, error) {
	roles, total, err := repository.GetRolePage(query)
	if err != nil {
		return nil, errs.SystemError("查询角色列表失败")
	}

	voList := make([]model.RolePageVO, len(roles))
	for i, role := range roles {
		voList[i] = model.RolePageVO{
			ID:         types.BigInt(role.ID),
			Name:       role.Name,
			Code:       role.Code,
			Sort:       role.Sort,
			Status:     role.Status,
			DataScope:  role.DataScope,
			CreateTime: role.CreateTime,
			UpdateTime: role.UpdateTime,
		}
	}

	return &common.PageResult{
		List:  voList,
		Total: total,
	}, nil
}

// GetRoleOptions 角色下拉选项
func GetRoleOptions() ([]common.Option[types.BigInt], error) {
	roles, err := repository.GetRoleOptions()
	if err != nil {
		return nil, errs.SystemError("查询角色选项失败")
	}

	options := make([]common.Option[types.BigInt], len(roles))
	for i, role := range roles {
		options[i] = common.Option[types.BigInt]{
			Value: types.BigInt(role.ID),
			Label: role.Name,
		}
	}

	return options, nil
}

// SaveRole 保存角色（新增或更新）
func SaveRole(form *model.RoleForm) error {
	exists, err := repository.CheckRoleNameExists(form.Name, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查角色名称失败")
	}
	if exists {
		return errs.BadRequest("角色名称已存在")
	}

	exists, err = repository.CheckRoleCodeExists(form.Code, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查角色编码失败")
	}
	if exists {
		return errs.BadRequest("角色编码已存在")
	}

	role := &model.Role{
		ID:        form.ID,
		Name:      form.Name,
		Code:      form.Code,
		Sort:      form.Sort,
		Status:    form.Status,
		DataScope: form.DataScope,
	}

	if form.ID == 0 {
		if err := repository.CreateRole(role); err != nil {
			return errs.SystemError("创建角色失败")
		}
		form.ID = role.ID
	} else {
		if err := repository.UpdateRole(role); err != nil {
			return errs.SystemError("更新角色失败")
		}
	}

	if len(form.MenuIds) > 0 {
		menuIds := make([]int64, len(form.MenuIds))
		for i, id := range form.MenuIds {
			menuIds[i] = int64(id)
		}
		if err := repository.UpdateRoleMenus(int64(form.ID), menuIds); err != nil {
			return errs.SystemError("更新角色菜单失败")
		}
	}

	return nil
}

// GetRoleForm 获取角色表单数据
func GetRoleForm(id int64) (*model.RoleForm, error) {
	role, err := repository.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("角色不存在")
		}
		return nil, errs.SystemError("查询角色失败")
	}

	menuIds, err := repository.GetRoleMenuIds(id)
	if err != nil {
		return nil, errs.SystemError("查询角色菜单失败")
	}

	menuIdsBigInt := make([]types.BigInt, len(menuIds))
	for i, id := range menuIds {
		menuIdsBigInt[i] = types.BigInt(id)
	}

	return &model.RoleForm{
		ID:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		Sort:      role.Sort,
		Status:    role.Status,
		DataScope: role.DataScope,
		MenuIds:   menuIdsBigInt,
	}, nil
}

// DeleteRole 删除角色
func DeleteRole(id int64) error {
	_, err := repository.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("角色不存在")
		}
		return errs.SystemError("查询角色失败")
	}

	if err := repository.DeleteRole(id); err != nil {
		return errs.SystemError("删除角色失败")
	}

	return nil
}

// GetRoleMenuIds 获取角色菜单ID列表
func GetRoleMenuIds(roleId int64) ([]int64, error) {
	menuIds, err := repository.GetRoleMenuIds(roleId)
	if err != nil {
		return nil, errs.SystemError("查询角色菜单失败")
	}
	return menuIds, nil
}

// UpdateRoleMenus 分配角色菜单权限
func UpdateRoleMenus(roleId int64, menuIds []int64) error {
	role, err := repository.GetRoleByID(roleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("角色不存在")
		}
		return errs.SystemError("查询角色失败")
	}

	if err := repository.UpdateRoleMenus(roleId, menuIds); err != nil {
		return errs.SystemError("更新角色菜单失败")
	}

	// 刷新该角色的权限缓存
	if err := RefreshRolePermsCacheByCode(role.Code); err != nil {
		// 日志记录但不阻断操作
		return errs.SystemError("刷新角色权限缓存失败")
	}

	return nil
}
