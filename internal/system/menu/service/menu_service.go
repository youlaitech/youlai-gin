package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/logger"
	"youlai-gin/internal/common/utils"
	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/repository"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Repository 菜单数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	List(ctx context.Context, query *model.MenuQuery) ([]model.Menu, error)
	Get(ctx context.Context, id int64) (*model.Menu, error)
	Create(ctx context.Context, menu *model.Menu) error
	Update(ctx context.Context, form *model.MenuForm) error
	Delete(ctx context.Context, id int64) error
	Options(ctx context.Context, onlyParent bool) ([]model.Menu, error)
	UserMenus(ctx context.Context, userId int64) ([]model.Menu, error)
	NameExists(ctx context.Context, name string, parentId int64, excludeId int64) (bool, error)
	RouteNameExists(ctx context.Context, routeName string, excludeId int64) (bool, error)
	ChildrenCount(ctx context.Context, parentId int64) (int64, error)
	MaxSortMenu(ctx context.Context, parentId int64) (*model.Menu, error)
	Children(ctx context.Context, parentId int64) ([]model.Menu, error)
	UpdateTreePathByParent(ctx context.Context, parentId int64, treePath string) error
	UpdateTreePath(ctx context.Context, id int64, treePath string) error
}

// RoleCacheRefresher 角色权限缓存刷新接口（依赖倒置：菜单定义接口，角色模块实现）
type RoleCacheRefresher interface {
	RefreshPermsCacheByMenus(menuIds []int64) error
}

// Service 菜单业务逻辑层
type Service struct {
	repo      Repository
	roleCache RoleCacheRefresher
}

// NewService 创建 Service 实例
func NewService(repo Repository, roleCache RoleCacheRefresher) *Service {
	return &Service{repo: repo, roleCache: roleCache}
}

// List 菜单列表（树形结构）
func (s *Service) List(ctx context.Context, query *model.MenuQuery) ([]*model.MenuVO, error) {
	menus, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询菜单列表失败")
	}

	menuVOs := make([]*model.MenuVO, len(menus))
	for i, menu := range menus {
		menuVOs[i] = &model.MenuVO{
			ID:          menu.ID,
			ParentID:    menu.ParentID,
			Name:        menu.Name,
			Type:        menu.Type,
			RouteName:   menu.RouteName,
			RoutePath:   menu.RoutePath,
			Component:   menu.Component,
			ExternalURL: menu.ExternalURL,
			Perm:        menu.Perm,
			AlwaysShow:  menu.AlwaysShow,
			KeepAlive:   menu.KeepAlive,
			Visible:     menu.Visible,
			Sort:        menu.Sort,
			Icon:        menu.Icon,
			Redirect:    menu.Redirect,
			CreateTime:  types.LocalTime(menu.CreateTime),
			UpdateTime:  types.LocalTime(menu.UpdateTime),
		}
	}

	return utils.BuildTreeSimple(
		menuVOs,
		func(m *model.MenuVO) int64 { return int64(m.ID) },
		func(m *model.MenuVO) int64 { return int64(m.ParentID) },
		func(m **model.MenuVO, children []*model.MenuVO) { (*m).Children = children },
	), nil
}

// Options 菜单下拉选项（树形结构）
func (s *Service) Options(ctx context.Context, onlyParent bool) ([]baseModel.Option[int64], error) {
	menus, err := s.repo.Options(ctx, onlyParent)
	if err != nil {
		return nil, errs.SystemError("查询菜单选项失败")
	}
	return s.buildOptions(0, menus), nil
}

// buildOptions 递归构建菜单选项树
func (s *Service) buildOptions(parentID int64, menus []model.Menu) []baseModel.Option[int64] {
	options := make([]baseModel.Option[int64], 0)
	for _, menu := range menus {
		if int64(menu.ParentID) != parentID {
			continue
		}
		option := baseModel.Option[int64]{Value: int64(menu.ID), Label: menu.Name}
		if children := s.buildOptions(int64(menu.ID), menus); len(children) > 0 {
			option.Children = children
		}
		options = append(options, option)
	}
	return options
}

// UserRoutes 获取用户路由（按菜单树生成，外链/内嵌特殊处理）
func (s *Service) UserRoutes(ctx context.Context, userId int64) ([]*model.RouteVO, error) {
	menus, err := s.repo.UserMenus(ctx, userId)
	if err != nil {
		return nil, errs.SystemError("查询用户菜单失败")
	}
	return s.buildRoutes(menus, 0), nil
}

// buildRoutes 递归构建路由树
func (s *Service) buildRoutes(menus []model.Menu, parentId int64) []*model.RouteVO {
	var routes []*model.RouteVO
	for _, menu := range menus {
		if int64(menu.ParentID) != parentId {
			continue
		}

		isExternal := menu.Type == "E"
		isEmbedded := isExternal && menu.Component == "iframe"

		path := menu.RoutePath
		if isExternal && !isEmbedded && menu.ExternalURL != "" {
			path = menu.ExternalURL
		}

		component := menu.Component
		if isEmbedded {
			component = "iframe"
		} else if isExternal {
			component = ""
		}

		meta := &model.RouteMeta{
			Title:      menu.Name,
			Icon:       menu.Icon,
			Hidden:     menu.Visible == 0,
			AlwaysShow: menu.AlwaysShow == 1,
			Params:     menu.Params,
		}
		if (menu.Type == "M" || isEmbedded) && menu.KeepAlive == 1 {
			meta.KeepAlive = true
		}
		if isEmbedded && menu.ExternalURL != "" {
			meta.ExternalURL = menu.ExternalURL
		}

		route := &model.RouteVO{
			Path:      path,
			Name:      menu.RouteName,
			Component: component,
			Redirect:  menu.Redirect,
			Meta:      meta,
		}
		if children := s.buildRoutes(menus, int64(menu.ID)); len(children) > 0 {
			route.Children = children
		}
		routes = append(routes, route)
	}
	return routes
}

// Create 新增菜单（按钮类菜单带权限标识时刷新角色权限缓存）
func (s *Service) Create(ctx context.Context, form *model.MenuForm) error {
	if err := s.prepareForm(ctx, form); err != nil {
		return err
	}

	treePath, err := s.generateTreePath(ctx, int64(form.ParentID))
	if err != nil {
		return errs.SystemError("查询父菜单失败")
	}

	menu := s.buildMenuEntity(form)
	menu.TreePath = treePath
	if err := s.repo.Create(ctx, menu); err != nil {
		return errs.SystemError("创建菜单失败")
	}
	form.ID = menu.ID

	if form.Type == "B" && form.Perm != "" {
		if err := s.refreshAffectedRolesCache([]int64{int64(menu.ID)}); err != nil {
			logger.Log.Sugar().Infof("刷新角色权限缓存失败: %v", err)
		}
	}

	return nil
}

// Update 更新菜单（重算树路径并级联子节点，刷新受影响角色的权限缓存）
func (s *Service) Update(ctx context.Context, id int64, form *model.MenuForm) error {
	form.ID = types.BigInt(id)

	if err := s.prepareForm(ctx, form); err != nil {
		return err
	}

	treePath, err := s.generateTreePath(ctx, int64(form.ParentID))
	if err != nil {
		return errs.SystemError("查询父菜单失败")
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新菜单失败")
	}
	if err := s.repo.UpdateTreePath(ctx, id, treePath); err != nil {
		logger.Log.Sugar().Infof("更新菜单树路径失败: %v", err)
	}

	if err := s.refreshAffectedRolesCache([]int64{id}); err != nil {
		logger.Log.Sugar().Infof("刷新角色权限缓存失败: %v", err)
	}

	s.updateChildrenTreePath(ctx, id, treePath)
	return nil
}

// prepareForm 保存前校验与规范化（父级自引用、同级名称、类型组件、路由名称）
func (s *Service) prepareForm(ctx context.Context, form *model.MenuForm) error {
	if form.ID != 0 && int64(form.ParentID) == int64(form.ID) {
		return errs.BadRequest("父级菜单不能为当前菜单")
	}

	exists, err := s.repo.NameExists(ctx, form.Name, int64(form.ParentID), int64(form.ID))
	if err != nil {
		return errs.SystemError("检查菜单名称失败")
	}
	if exists {
		return errs.BadRequest("同级菜单名称已存在")
	}

	isExternal := form.Type == "E"
	isEmbedded := isExternal && form.Component == "iframe"
	if form.Type == "C" {
		if form.ParentID == 0 && form.RoutePath != "" && !strings.HasPrefix(form.RoutePath, "/") {
			form.RoutePath = "/" + form.RoutePath
		}
		form.Component = "Layout"
	} else if isExternal && !isEmbedded {
		form.Component = ""
	}

	if form.Type == "M" || isEmbedded {
		if form.RouteName != "" {
			dup, err := s.repo.RouteNameExists(ctx, form.RouteName, int64(form.ID))
			if err != nil {
				return errs.SystemError("检查路由名称失败")
			}
			if dup {
				return errs.BadRequest("路由名称已存在")
			}
		}
	} else {
		form.RouteName = ""
	}

	return nil
}

// generateTreePath 根据父 ID 生成树路径：根节点返回 "0"，非根节点返回父树路径 + "," + 父ID
func (s *Service) generateTreePath(ctx context.Context, parentId int64) (string, error) {
	if parentId == 0 {
		return "0", nil
	}
	parent, err := s.repo.Get(ctx, parentId)
	if err != nil {
		return "", err
	}
	return parent.TreePath + "," + fmt.Sprintf("%d", parent.ID), nil
}

// updateChildrenTreePath 递归更新子菜单树路径
func (s *Service) updateChildrenTreePath(ctx context.Context, id int64, treePath string) {
	children, err := s.repo.Children(ctx, id)
	if err != nil || len(children) == 0 {
		return
	}

	childTreePath := treePath + "," + fmt.Sprintf("%d", id)
	if err := s.repo.UpdateTreePathByParent(ctx, id, childTreePath); err != nil {
		logger.Log.Sugar().Infof("更新子节点树路径失败(parent=%d): %v", id, err)
		return
	}

	for _, child := range children {
		s.updateChildrenTreePath(ctx, int64(child.ID), childTreePath)
	}
}

// buildMenuEntity 由表单构造新增实体
func (s *Service) buildMenuEntity(form *model.MenuForm) *model.Menu {
	return &model.Menu{
		ID:          form.ID,
		ParentID:    form.ParentID,
		Name:        form.Name,
		Type:        form.Type,
		RouteName:   form.RouteName,
		RoutePath:   form.RoutePath,
		Component:   form.Component,
		ExternalURL: form.ExternalURL,
		Perm:        form.Perm,
		AlwaysShow:  form.AlwaysShow,
		KeepAlive:   form.KeepAlive,
		Visible:     form.Visible,
		Sort:        form.Sort,
		Icon:        form.Icon,
		Redirect:    form.Redirect,
		Params:      s.keyValueToMap(form.Params),
	}
}

// GetForm 获取菜单表单数据
func (s *Service) GetForm(ctx context.Context, id int64) (*model.MenuForm, error) {
	menu, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("菜单不存在")
		}
		return nil, errs.SystemError("查询菜单失败")
	}

	return &model.MenuForm{
		ID:          menu.ID,
		ParentID:    menu.ParentID,
		Name:        menu.Name,
		Type:        menu.Type,
		RouteName:   menu.RouteName,
		RoutePath:   menu.RoutePath,
		Component:   menu.Component,
		ExternalURL: menu.ExternalURL,
		Perm:        menu.Perm,
		AlwaysShow:  menu.AlwaysShow,
		KeepAlive:   menu.KeepAlive,
		Visible:     menu.Visible,
		Sort:        menu.Sort,
		Icon:        menu.Icon,
		Redirect:    menu.Redirect,
		Params:      s.mapToKeyValue(menu.Params),
	}, nil
}

// Delete 删除菜单（存在子菜单时拒绝；按钮类菜单删除后刷新角色权限缓存）
func (s *Service) Delete(ctx context.Context, id int64) error {
	menu, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("菜单不存在")
		}
		return errs.SystemError("查询菜单失败")
	}

	count, err := s.repo.ChildrenCount(ctx, id)
	if err != nil {
		return errs.SystemError("查询子菜单失败")
	}
	if count > 0 {
		return errs.BadRequest("请先删除子菜单")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errs.SystemError("删除菜单失败")
	}

	if menu.Type == "B" && menu.Perm != "" {
		if err := s.refreshAffectedRolesCache([]int64{id}); err != nil {
			logger.Log.Sugar().Infof("刷新角色权限缓存失败: %v", err)
		}
	}

	return nil
}

// refreshAffectedRolesCache 刷新受影响角色的权限缓存
func (s *Service) refreshAffectedRolesCache(menuIds []int64) error {
	return s.roleCache.RefreshPermsCacheByMenus(menuIds)
}

// AddMenuForCodegen 代码生成时追加菜单及按钮权限
func (s *Service) AddMenuForCodegen(parentMenuId int64, tableName, moduleName, businessName, entityName string) error {
	ctx := context.Background()

	parentMenu, err := s.repo.Get(ctx, parentMenuId)
	if err != nil {
		return errs.NotFound("父菜单不存在")
	}

	sort := 1
	if maxSortMenu, err := s.repo.MaxSortMenu(ctx, parentMenuId); err == nil && maxSortMenu != nil {
		sort = int(maxSortMenu.Sort) + 1
	}

	entityKebab := s.toKebabCase(entityName)
	menu := &model.Menu{
		ParentID:  types.BigInt(parentMenuId),
		Name:      businessName,
		Type:      "M",
		RouteName: entityName,
		RoutePath: entityKebab,
		Component: moduleName + "/" + entityKebab + "/index",
		Sort:      sort,
		Visible:   1,
		TreePath:  fmt.Sprintf("%s,%d", parentMenu.TreePath, parentMenuId),
	}
	if err := s.repo.Create(ctx, menu); err != nil {
		return errs.SystemError("创建菜单失败")
	}

	permPrefix := moduleName + ":" + strings.ReplaceAll(tableName, "_", "-") + ":"
	for i, action := range []string{"查询", "新增", "修改", "删除"} {
		perm := permPrefix + []string{"list", "create", "update", "delete"}[i]
		button := &model.Menu{
			ParentID: types.BigInt(menu.ID),
			Type:     "B",
			Name:     action,
			Perm:     perm,
			Sort:     i + 1,
			TreePath: fmt.Sprintf("%s,%d", menu.TreePath, menu.ID),
		}
		if err := s.repo.Create(ctx, button); err != nil {
			logger.Log.Sugar().Infof("创建按钮菜单失败: %v", err)
		}
	}

	return nil
}

// keyValueToMap 键值对列表转 map（空列表返回 nil）
func (s *Service) keyValueToMap(kvs []baseModel.KeyValue) map[string]any {
	m := make(map[string]any)
	for _, kv := range kvs {
		m[kv.Key] = kv.Value
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// mapToKeyValue map 转键值对列表
func (s *Service) mapToKeyValue(m map[string]any) []baseModel.KeyValue {
	kvs := make([]baseModel.KeyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, baseModel.KeyValue{Key: k, Value: fmt.Sprintf("%v", v)})
	}
	return kvs
}

// toKebabCase 驼峰转短横线命名
func (s *Service) toKebabCase(str string) string {
	if str == "" {
		return ""
	}
	var result []rune
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, []rune(strings.ToLower(string(r)))...)
	}
	return string(result)
}

// AddMenuForCodegen 包级桥接（codegen 模块未迁移，走全局 DB 装配）
func AddMenuForCodegen(parentMenuId int64, tableName, moduleName, businessName, entityName string) error {
	return NewService(repository.NewRepository(database.DB), nil).AddMenuForCodegen(parentMenuId, tableName, moduleName, businessName, entityName)
}
