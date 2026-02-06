package service

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/repository"
	roleRepo "youlai-gin/internal/system/role/repository"
	roleService "youlai-gin/internal/system/role/service"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/utils"
)

// GetMenuList 菜单列表（树形结构）
func GetMenuList(query *model.MenuQuery) ([]*model.MenuVO, error) {
	menus, err := repository.GetMenuList(query)
	if err != nil {
		return nil, errs.SystemError("查询菜单列表失败")
	}

	menuVOs := make([]*model.MenuVO, len(menus))
	for i, menu := range menus {
		menuVOs[i] = &model.MenuVO{
			ID:         menu.ID,
			ParentID:   menu.ParentID,
			Name:       menu.Name,
			Type:       menu.Type,
			RouteName:  menu.RouteName,
			RoutePath:  menu.RoutePath,
			Component:  menu.Component,
			Perm:       menu.Perm,
			AlwaysShow: menu.AlwaysShow,
			KeepAlive:  menu.KeepAlive,
			Visible:    menu.Visible,
			Sort:       menu.Sort,
			Icon:       menu.Icon,
			Redirect:   menu.Redirect,
			CreateTime: types.LocalTime(menu.CreateTime),
			UpdateTime: types.LocalTime(menu.UpdateTime),
		}
	}

	tree := utils.BuildTreeSimple(
		menuVOs,
		func(m *model.MenuVO) int64 { return int64(m.ID) },
		func(m *model.MenuVO) int64 { return int64(m.ParentID) },
		func(m **model.MenuVO, children []*model.MenuVO) {
			(*m).Children = children
		},
	)

	return tree, nil
}

// GetMenuOptions 菜单下拉选项
func GetMenuOptions(onlyParent bool) ([]common.Option[int64], error) {
	menus, err := repository.GetMenuOptions(onlyParent)
	if err != nil {
		return nil, errs.SystemError("查询菜单选项失败")
	}

	options := buildMenuOptions(0, menus)
	return options, nil
}

func buildMenuOptions(parentID int64, menus []model.Menu) []common.Option[int64] {
	options := make([]common.Option[int64], 0)
	for _, menu := range menus {
		if int64(menu.ParentID) != parentID {
			continue
		}

		option := common.Option[int64]{
			Value: int64(menu.ID),
			Label: menu.Name,
		}
		children := buildMenuOptions(int64(menu.ID), menus)
		if len(children) > 0 {
			option.Children = children
		}
		options = append(options, option)
	}

	return options
}

// GetCurrentUserRoutes 获取当前用户路由
func GetCurrentUserRoutes(userId int64) ([]*model.RouteVO, error) {
	menus, err := repository.GetUserMenus(userId)
	if err != nil {
		return nil, errs.SystemError("查询用户菜单失败")
	}

	routes := buildRoutes(menus, 0)
	return routes, nil
}

// buildRoutes 递归构建路由树
func buildRoutes(menus []model.Menu, parentId int64) []*model.RouteVO {
	var routes []*model.RouteVO

	for _, menu := range menus {
		if int64(menu.ParentID) == parentId {
			route := &model.RouteVO{
				Path:      menu.RoutePath,
				Name:      menu.RouteName,
				Component: menu.Component,
				Redirect:  menu.Redirect,
				Meta: &model.RouteMeta{
					Title:      menu.Name,
					Icon:       menu.Icon,
					Hidden:     menu.Visible == 0,
					AlwaysShow: menu.AlwaysShow == 1,
					KeepAlive:  menu.KeepAlive == 1,
					Params:     menu.Params,
				},
			}

			children := buildRoutes(menus, int64(menu.ID))
			if len(children) > 0 {
				route.Children = children
			}

			routes = append(routes, route)
		}
	}

	return routes
}

// SaveMenu 保存菜单（新增或更新）
func SaveMenu(form *model.MenuForm) error {
	exists, err := repository.CheckMenuNameExists(form.Name, int64(form.ParentID), int64(form.ID))
	if err != nil {
		return errs.SystemError("检查菜单名称失败")
	}
	if exists {
		return errs.BadRequest("同级菜单名称已存在")
	}

	menu := &model.Menu{
		ID:         form.ID,
		ParentID:   form.ParentID,
		Name:       form.Name,
		Type:       form.Type,
		RouteName:  form.RouteName,
		RoutePath:  form.RoutePath,
		Component:  form.Component,
		Perm:       form.Perm,
		AlwaysShow: form.AlwaysShow,
		KeepAlive:  form.KeepAlive,
		Visible:    form.Visible,
		Sort:       form.Sort,
		Icon:       form.Icon,
		Redirect:   form.Redirect,
		Params:     form.Params,
	}

	if form.ParentID == 0 {
		menu.TreePath = "0"
	} else {
		parent, err := repository.GetMenuByID(int64(form.ParentID))
		if err != nil {
			return errs.SystemError("查询父菜单失败")
		}
		menu.TreePath = fmt.Sprintf("%s,%d", parent.TreePath, parent.ID)
	}

	// 保存菜单
	var menuID int64
	isUpdate := form.ID > 0

	if !isUpdate {
		if err := repository.CreateMenu(menu); err != nil {
			return errs.SystemError("创建菜单失败")
		}
		menuID = int64(menu.ID)
	} else {
		if err := repository.UpdateMenu(menu); err != nil {
			return errs.SystemError("更新菜单失败")
		}
		menuID = int64(menu.ID)
	}

	// 刷新受影响角色的权限缓存
	if menu.Type == "B" && menu.Perm != "" {
		if err := refreshAffectedRolesCache([]int64{menuID}); err != nil {
			log.Printf("刷新角色权限缓存失败: %v", err)
		}
	}

	return nil
}

// GetMenuForm 获取菜单表单数据
func GetMenuForm(id int64) (*model.MenuForm, error) {
	menu, err := repository.GetMenuByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("菜单不存在")
		}
		return nil, errs.SystemError("查询菜单失败")
	}

	return &model.MenuForm{
		ID:         menu.ID,
		ParentID:   menu.ParentID,
		Name:       menu.Name,
		Type:       menu.Type,
		RouteName:  menu.RouteName,
		RoutePath:  menu.RoutePath,
		Component:  menu.Component,
		Perm:       menu.Perm,
		AlwaysShow: menu.AlwaysShow,
		KeepAlive:  menu.KeepAlive,
		Visible:    menu.Visible,
		Sort:       menu.Sort,
		Icon:       menu.Icon,
		Redirect:   menu.Redirect,
		Params:     menu.Params,
	}, nil
}

// DeleteMenu 删除菜单
func DeleteMenu(id int64) error {
	menu, err := repository.GetMenuByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("菜单不存在")
		}
		return errs.SystemError("查询菜单失败")
	}

	count, err := repository.GetChildrenCount(id)
	if err != nil {
		return errs.SystemError("查询子菜单失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除子菜单")
	}

	if err := repository.DeleteMenu(id); err != nil {
		return errs.SystemError("删除菜单失败")
	}

	// 刷新受影响角色的权限缓存
	// 仅当删除的是按钮且有权限标识时才刷新
	if menu.Type == "B" && menu.Perm != "" {
		if err := refreshAffectedRolesCache([]int64{id}); err != nil {
			log.Printf("刷新角色权限缓存失败: %v", err)
			// 不阻断操作，记录日志即可
		}
	}

	return nil
}

// refreshAffectedRolesCache 刷新受菜单影响的角色权限缓存（内部辅助函数）
func refreshAffectedRolesCache(menuIds []int64) error {
	if len(menuIds) == 0 {
		return nil
	}

	// 查询受影响的角色
	roleCodes, err := roleRepo.GetRolesAffectedByMenus(menuIds)
	if err != nil {
		return fmt.Errorf("查询受影响的角色失败: %w", err)
	}

	if len(roleCodes) == 0 {
		// 没有角色受影响，无需刷新
		return nil
	}

	// 批量刷新这些角色的权限缓存
	if err := roleService.RefreshRolePermsCacheByCodes(roleCodes); err != nil {
		return fmt.Errorf("批量刷新角色权限缓存失败: %w", err)
	}

	return nil
}

// GetUserPermissions 获取用户按钮权限
func GetUserPermissions(userId int64) ([]string, error) {
	perms, err := repository.GetUserButtonPerms(userId)
	if err != nil {
		return nil, errs.SystemError("查询用户权限失败")
	}
	return perms, nil
}
