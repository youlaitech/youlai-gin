package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	permService "youlai-gin/internal/system/permission/service"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/internal/system/role/repository"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
	"youlai-gin/pkg/types"
)

// GetRolePage 角色分页列表
func GetRolePage(query *model.RoleQuery) (*common.PagedData, error) {
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
			CreateTime: types.LocalTime(role.CreateTime),
			UpdateTime: types.LocalTime(role.UpdateTime),
		}
	}

	return &common.PagedData{List: voList, Total: total}, nil
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
	var oldDataScope int
	if form.ID != 0 {
		oldRole, err := repository.GetRoleByID(int64(form.ID))
		if err == nil && oldRole != nil {
			oldDataScope = oldRole.DataScope
		}
	}

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

	// 数据权限发生变化时，失效该角色关联用户的登录态（JWT tokenVersion）
	if form.ID != 0 && oldDataScope != 0 && oldDataScope != form.DataScope {
		userIds, err := userRepo.ListUserIDsByRoleID(int64(form.ID))
		if err == nil && len(userIds) > 0 {
			for _, uid := range userIds {
				_ = invalidateUserSessions(uid)
			}
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

	// 自定义数据权限：同步维护 sys_role_dept（并在变更时失效会话）
	roleId := int64(form.ID)
	if form.DataScope == permService.DataScopeCustom {
		deptIds := make([]int64, 0, len(form.DeptIds))
		for _, id := range form.DeptIds {
			if int64(id) > 0 {
				deptIds = append(deptIds, int64(id))
			}
		}
		if err := UpdateRoleDepts(roleId, deptIds); err != nil {
			return err
		}
	} else {
		// 非自定义：清理历史自定义部门
		if err := repository.UpdateRoleDepts(roleId, nil); err != nil {
			return errs.SystemError("清理角色自定义部门失败")
		}
	}

	return nil
}

func invalidateUserSessions(userId int64) error {
	if userId <= 0 {
		return nil
	}
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", redis.UserTokenVersion, userId)
	return redis.Client.Incr(ctx, key).Err()
}

// GetRoleDeptIds 获取角色自定义部门ID列表
func GetRoleDeptIds(roleId int64) ([]int64, error) {
	if roleId <= 0 {
		return []int64{}, nil
	}
	return repository.GetRoleDeptIds(roleId)
}

// UpdateRoleDepts 更新角色自定义部门，并失效该角色关联用户会话
func UpdateRoleDepts(roleId int64, deptIds []int64) error {
	if roleId <= 0 {
		return errs.BadRequest("无效的角色ID")
	}

	oldDeptIds, _ := repository.GetRoleDeptIds(roleId)
	if err := repository.UpdateRoleDepts(roleId, deptIds); err != nil {
		return errs.SystemError("更新角色自定义部门失败")
	}

	oldSet := make(map[int64]struct{}, len(oldDeptIds))
	for _, id := range oldDeptIds {
		oldSet[id] = struct{}{}
	}
	newSet := make(map[int64]struct{}, len(deptIds))
	for _, id := range deptIds {
		if id > 0 {
			newSet[id] = struct{}{}
		}
	}
	changed := len(oldSet) != len(newSet)
	if !changed {
		for id := range oldSet {
			if _, ok := newSet[id]; !ok {
				changed = true
				break
			}
		}
	}

	if changed {
		userIds, err := userRepo.ListUserIDsByRoleID(roleId)
		if err == nil && len(userIds) > 0 {
			for _, uid := range userIds {
				_ = invalidateUserSessions(uid)
			}
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

	var deptIdsBigInt []types.BigInt
	if role.DataScope == permService.DataScopeCustom {
		deptIds, err := repository.GetRoleDeptIds(id)
		if err != nil {
			return nil, errs.SystemError("查询角色自定义部门失败")
		}
		deptIdsBigInt = make([]types.BigInt, len(deptIds))
		for i, did := range deptIds {
			deptIdsBigInt[i] = types.BigInt(did)
		}
	}

	return &model.RoleForm{
		ID:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		Sort:      role.Sort,
		Status:    role.Status,
		DataScope: role.DataScope,
		DeptIds:   deptIdsBigInt,
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
