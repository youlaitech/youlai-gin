package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/logger"
	"youlai-gin/internal/common/utils"
	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/repository"
	roleRepo "youlai-gin/internal/system/role/repository"
	roleService "youlai-gin/internal/system/role/service"
	common "youlai-gin/pkg/model"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// Service 菜单业务逻辑层
type Service struct {
	repo *repository.Repository
}

// NewService 创建 Service 实例
func NewService(repo *repository.Repository) *Service { return &Service{repo: repo} }

// GetMenuList 菜单列表（树形结构）
func (s *Service) GetMenuList(query *model.MenuQuery) ([]*model.MenuVO, error) {
	menus, err := s.repo.GetMenuList(query)
	if err != nil { return nil, errs.SystemError("查询菜单列表失败") }
	menuVOs := make([]*model.MenuVO, len(menus))
	for i, menu := range menus {
		menuVOs[i] = &model.MenuVO{ID: menu.ID, ParentID: menu.ParentID, Name: menu.Name, Type: menu.Type, RouteName: menu.RouteName, RoutePath: menu.RoutePath, Component: menu.Component, Perm: menu.Perm, AlwaysShow: menu.AlwaysShow, KeepAlive: menu.KeepAlive, Visible: menu.Visible, Sort: menu.Sort, Icon: menu.Icon, Redirect: menu.Redirect, CreateTime: types.LocalTime(menu.CreateTime), UpdateTime: types.LocalTime(menu.UpdateTime)}
	}
	return utils.BuildTreeSimple(menuVOs, func(m *model.MenuVO) int64 { return int64(m.ID) }, func(m *model.MenuVO) int64 { return int64(m.ParentID) }, func(m **model.MenuVO, children []*model.MenuVO) { (*m).Children = children }), nil
}

// GetMenuOptions 菜单下拉选项
func (s *Service) GetMenuOptions(onlyParent bool) ([]common.Option[int64], error) {
	menus, err := s.repo.GetMenuOptions(onlyParent)
	if err != nil { return nil, errs.SystemError("查询菜单选项失败") }
	return s.buildMenuOptions(0, menus), nil
}

func (s *Service) buildMenuOptions(parentID int64, menus []model.Menu) []common.Option[int64] {
	options := make([]common.Option[int64], 0)
	for _, menu := range menus {
		if int64(menu.ParentID) != parentID { continue }
		option := common.Option[int64]{Value: int64(menu.ID), Label: menu.Name}
		if children := s.buildMenuOptions(int64(menu.ID), menus); len(children) > 0 { option.Children = children }
		options = append(options, option)
	}
	return options
}

// GetCurrentUserRoutes 获取当前用户路由
func (s *Service) GetCurrentUserRoutes(userId int64) ([]*model.RouteVO, error) {
	menus, err := s.repo.GetUserMenus(userId)
	if err != nil { return nil, errs.SystemError("查询用户菜单失败") }
	return s.buildRoutes(menus, 0), nil
}

func (s *Service) buildRoutes(menus []model.Menu, parentId int64) []*model.RouteVO {
	var routes []*model.RouteVO
	for _, menu := range menus {
		if int64(menu.ParentID) == parentId {
			route := &model.RouteVO{Path: menu.RoutePath, Name: menu.RouteName, Component: menu.Component, Redirect: menu.Redirect, Meta: &model.RouteMeta{Title: menu.Name, Icon: menu.Icon, Hidden: menu.Visible == 0, AlwaysShow: menu.AlwaysShow == 1, KeepAlive: menu.KeepAlive == 1, Params: menu.Params}}
			if children := s.buildRoutes(menus, int64(menu.ID)); len(children) > 0 { route.Children = children }
			routes = append(routes, route)
		}
	}
	return routes
}

// SaveMenu 保存菜单（新增或更新）
func (s *Service) SaveMenu(form *model.MenuForm) error {
	exists, err := s.repo.CheckMenuNameExists(form.Name, int64(form.ParentID), int64(form.ID))
	if err != nil { return errs.SystemError("检查菜单名称失败") }
	if exists { return errs.BadRequest("同级菜单名称已存在") }
	menu := &model.Menu{ID: form.ID, ParentID: form.ParentID, Name: form.Name, Type: form.Type, RouteName: form.RouteName, RoutePath: form.RoutePath, Component: form.Component, Perm: form.Perm, AlwaysShow: form.AlwaysShow, KeepAlive: form.KeepAlive, Visible: form.Visible, Sort: form.Sort, Icon: form.Icon, Redirect: form.Redirect, Params: s.keyValueToMap(form.Params)}
	if form.ParentID == 0 { menu.TreePath = "0" } else {
		parent, err := s.repo.GetMenuByID(int64(form.ParentID))
		if err != nil { return errs.SystemError("查询父菜单失败") }
		menu.TreePath = fmt.Sprintf("%s,%d", parent.TreePath, parent.ID)
	}
	var menuID int64
	if form.ID == 0 { if err := s.repo.CreateMenu(menu); err != nil { return errs.SystemError("创建菜单失败") }; menuID = int64(menu.ID) } else { if err := s.repo.UpdateMenu(menu); err != nil { return errs.SystemError("更新菜单失败") }; menuID = int64(menu.ID) }
	if menu.Type == "B" && menu.Perm != "" { if err := s.refreshAffectedRolesCache([]int64{menuID}); err != nil { logger.Log.Sugar().Infof("刷新角色权限缓存失败: %v", err) } }
	return nil
}

// GetMenuForm 获取菜单表单数据
func (s *Service) GetMenuForm(id int64) (*model.MenuForm, error) {
	menu, err := s.repo.GetMenuByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, errs.NotFound("菜单不存在") }
		return nil, errs.SystemError("查询菜单失败")
	}
	return &model.MenuForm{ID: menu.ID, ParentID: menu.ParentID, Name: menu.Name, Type: menu.Type, RouteName: menu.RouteName, RoutePath: menu.RoutePath, Component: menu.Component, Perm: menu.Perm, AlwaysShow: menu.AlwaysShow, KeepAlive: menu.KeepAlive, Visible: menu.Visible, Sort: menu.Sort, Icon: menu.Icon, Redirect: menu.Redirect, Params: s.mapToKeyValue(menu.Params)}, nil
}

// DeleteMenu 删除菜单
func (s *Service) DeleteMenu(id int64) error {
	menu, err := s.repo.GetMenuByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errs.NotFound("菜单不存在") }
		return errs.SystemError("查询菜单失败")
	}
	if count, err := s.repo.GetChildrenCount(id); err != nil { return errs.SystemError("查询子菜单失败") } else if count > 0 { return errs.BadRequest("请先删除子菜单") }
	if err := s.repo.DeleteMenu(id); err != nil { return errs.SystemError("删除菜单失败") }
	if menu.Type == "B" && menu.Perm != "" { if err := s.refreshAffectedRolesCache([]int64{id}); err != nil { logger.Log.Sugar().Infof("刷新角色权限缓存失败: %v", err) } }
	return nil
}

func (s *Service) refreshAffectedRolesCache(menuIds []int64) error {
	if len(menuIds) == 0 { return nil }
	roleCodes, err := roleRepo.GetRolesAffectedByMenus(menuIds)
	if err != nil { return fmt.Errorf("查询受影响的角色失败: %w", err) }
	if len(roleCodes) == 0 { return nil }
	if err := roleService.RefreshRolePermsCacheByCodes(roleCodes); err != nil { return fmt.Errorf("批量刷新角色权限缓存失败: %w", err) }
	return nil
}

// AddMenuForCodegen 代码生成时添加菜单
func (s *Service) AddMenuForCodegen(parentMenuId int64, tableName, moduleName, businessName, entityName string) error {
	parentMenu, err := s.repo.GetMenuByID(parentMenuId)
	if err != nil { return errs.NotFound("父菜单不存在") }
	sort := 1
	if maxSortMenu, err := s.repo.GetMaxSortMenuByParentID(parentMenuId); err == nil && maxSortMenu != nil { sort = int(maxSortMenu.Sort) + 1 }
	entityKebab := s.toKebabCase(entityName)
	menu := &model.Menu{ParentID: types.BigInt(parentMenuId), Name: businessName, Type: "M", RouteName: entityName, RoutePath: entityKebab, Component: moduleName + "/" + entityKebab + "/index", Sort: sort, Visible: 1, TreePath: fmt.Sprintf("%s,%d", parentMenu.TreePath, parentMenuId)}
	if err := s.repo.CreateMenu(menu); err != nil { return errs.SystemError("创建菜单失败") }
	permPrefix := moduleName + ":" + strings.ReplaceAll(tableName, "_", "-") + ":"
	for i, action := range []string{"查询", "新增", "修改", "删除"} {
		perm := permPrefix + []string{"list", "create", "update", "delete"}[i]
		button := &model.Menu{ParentID: types.BigInt(menu.ID), Type: "B", Name: action, Perm: perm, Sort: i + 1, TreePath: fmt.Sprintf("%s,%d", menu.TreePath, menu.ID)}
		if err := s.repo.CreateMenu(button); err != nil { logger.Log.Sugar().Infof("创建按钮菜单失败: %v", err) }
	}
	return nil
}

func (s *Service) keyValueToMap(kvs []common.KeyValue) map[string]any { m := make(map[string]any); for _, kv := range kvs { m[kv.Key] = kv.Value }; if len(m) == 0 { return nil }; return m }
func (s *Service) mapToKeyValue(m map[string]any) []common.KeyValue { kvs := make([]common.KeyValue, 0, len(m)); for k, v := range m { kvs = append(kvs, common.KeyValue{Key: k, Value: fmt.Sprintf("%v", v)}) }; return kvs }
func (s *Service) toKebabCase(str string) string { if str == "" { return "" }; var result []rune; for i, r := range str { if i > 0 && 'A' <= r && r <= 'Z' { result = append(result, '-') }; result = append(result, []rune(strings.ToLower(string(r)))...) }; return string(result) }

// AddMenuForCodegen 包级桥接（codegen 调用）
func AddMenuForCodegen(parentMenuId int64, tableName, moduleName, businessName, entityName string) error {
	return (&Service{repo: repository.NewRepository(database.DB)}).AddMenuForCodegen(parentMenuId, tableName, moduleName, businessName, entityName)
}
