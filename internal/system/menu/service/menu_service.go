package service

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"youlai-gin/internal/common/ai"
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
	Options(ctx context.Context, types []string) ([]model.Menu, error)
	UserMenus(ctx context.Context, userId int64) ([]model.Menu, error)
	NameExists(ctx context.Context, name string, parentId int64, excludeId int64) (bool, error)
	RouteNameExists(ctx context.Context, routeName string, excludeId int64) (bool, error)
	ChildrenCount(ctx context.Context, parentId int64) (int64, error)
	MaxSortMenu(ctx context.Context, parentId int64) (*model.Menu, error)
	Children(ctx context.Context, parentId int64) ([]model.Menu, error)
	Siblings(ctx context.Context, parentId int64, limit int) ([]model.Menu, error)
	DeleteWithDescendants(ctx context.Context, id int64) error
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

//go:embed prompts/menu/system.md
var menuSystemPrompt string

// siblingSampleLimit 同级菜单取样条数，样本只用于让模型沿用既有命名风格
const siblingSampleLimit = 10

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

// Options 菜单下拉选项（树形结构），types 非空时按菜单类型过滤
func (s *Service) Options(ctx context.Context, types []string) ([]baseModel.Option[int64], error) {
	menus, err := s.repo.Options(ctx, types)
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
			Title:  menu.Name,
			Icon:   menu.Icon,
			Hidden: menu.Visible == 0,
			Params: menu.Params,
		}
		if (menu.Type == "M" || isEmbedded) && menu.KeepAlive == 1 {
			meta.KeepAlive = true
		}
		if isExternal && menu.ExternalURL != "" {
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
	// 父级层级校验：按钮只能挂在菜单下，其他类型只能挂在顶级或目录下
	if err := s.validateMenuParent(ctx, form); err != nil {
		return err
	}

	// 新增菜单未指定排序时排到同级末尾
	if form.Sort == 0 {
		form.Sort = s.resolveNextSort(ctx, int64(form.ParentID))
	}

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

	// 新增页面菜单时按需生成增删改查按钮
	if form.GenerateCrudButtons && strings.TrimSpace(form.ButtonPermPrefix) != "" {
		s.saveCrudButtons(ctx, int64(menu.ID), strings.TrimSpace(form.ButtonPermPrefix), menu.TreePath)
	}

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

	// 父级层级校验：父级变更时才校验，未变更时放行历史遗留的非法层级
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.BadRequest("菜单不存在")
		}
		return errs.SystemError("查询菜单失败")
	}
	if int64(current.ParentID) != int64(form.ParentID) {
		if err := s.validateMenuParent(ctx, form); err != nil {
			return err
		}
	}

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
		KeepAlive:   menu.KeepAlive,
		Visible:     menu.Visible,
		Sort:        menu.Sort,
		Icon:        menu.Icon,
		Redirect:    menu.Redirect,
		Params:      s.mapToKeyValue(menu.Params),
	}, nil
}

// Delete 删除菜单（级联删除子菜单并解绑角色关联；按钮类菜单删除后刷新角色权限缓存）
func (s *Service) Delete(ctx context.Context, id int64) error {
	menu, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("菜单不存在")
		}
		return errs.SystemError("查询菜单失败")
	}

	if err := s.repo.DeleteWithDescendants(ctx, id); err != nil {
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

// validateMenuParent 校验上级菜单层级：按钮只能挂在菜单下，其他类型只能挂在顶级或目录下
func (s *Service) validateMenuParent(ctx context.Context, form *model.MenuForm) error {
	parentID := int64(form.ParentID)
	isButton := form.Type == "B"

	if parentID == 0 {
		if isButton {
			return errs.BadRequest("按钮权限只能挂在菜单下")
		}
		return nil
	}

	parent, err := s.repo.Get(ctx, parentID)
	if err != nil {
		return errs.NotFound("上级菜单不存在")
	}

	parentIsMenu := parent.Type == "M"
	if isButton && !parentIsMenu {
		return errs.BadRequest("按钮权限只能挂在菜单下")
	}
	if !isButton && parentIsMenu {
		return errs.BadRequest("菜单下只能挂按钮权限")
	}
	return nil
}

// resolveNextSort 同级菜单的最大排序 + 1
func (s *Service) resolveNextSort(ctx context.Context, parentID int64) int {
	if maxSortMenu, err := s.repo.MaxSortMenu(ctx, parentID); err == nil && maxSortMenu != nil {
		return int(maxSortMenu.Sort) + 1
	}
	return 1
}

// saveCrudButtons 生成增删改查按钮权限
func (s *Service) saveCrudButtons(ctx context.Context, menuID int64, permPrefix, menuTreePath string) {
	names := []string{"查询", "新增", "修改", "删除"}
	actions := []string{"list", "create", "update", "delete"}
	// 树路径记录祖先链，等于所属菜单的树路径加上所属菜单ID，删除菜单时据此级联
	buttonTreePath := fmt.Sprintf("%s,%d", menuTreePath, menuID)

	for i, name := range names {
		button := &model.Menu{
			ParentID: types.BigInt(menuID),
			Type:     "B",
			Name:     name,
			Perm:     permPrefix + ":" + actions[i],
			Sort:     i + 1,
			Visible:  1,
			TreePath: buttonTreePath,
		}
		if err := s.repo.Create(ctx, button); err != nil {
			logger.Log.Sugar().Infof("创建 CRUD 按钮菜单失败: %v", err)
		}
	}
}

// ── AI ────────────────────────────────────────────────────────

// AiFill 用 AI 推断菜单的访问路径、权限标识与图标，命名风格参照同级菜单
func (s *Service) AiFill(ctx context.Context, form *model.MenuAiFillForm) (*model.MenuAiFillVO, error) {
	userPrompt, err := s.buildAiUserPrompt(ctx, form)
	if err != nil {
		return nil, err
	}

	var vo model.MenuAiFillVO
	if err := ai.ChatJSON(ctx, menuSystemPrompt, userPrompt, &vo); err != nil {
		return nil, err
	}
	if vo.IconKeywords == nil {
		vo.IconKeywords = []string{}
	}
	return &vo, nil
}

// buildAiUserPrompt 构造用户提示词，附带同级菜单样本供模型沿用既有命名风格
func (s *Service) buildAiUserPrompt(ctx context.Context, form *model.MenuAiFillForm) (string, error) {
	parentID := int64(form.ParentID)

	parentName := "顶级"
	if parentID > 0 {
		if parent, err := s.repo.Get(ctx, parentID); err == nil && parent != nil {
			parentName = parent.Name
		}
	}

	siblings, err := s.repo.Siblings(ctx, parentID, siblingSampleLimit)
	if err != nil {
		return "", errs.SystemError("查询同级菜单失败")
	}

	var builder strings.Builder
	builder.WriteString("菜单名称：" + form.Name + "\n")
	builder.WriteString("菜单类型：" + menuTypeLabel(form.Type) + "\n")
	builder.WriteString("上级菜单：" + parentName + "\n")
	builder.WriteString("同级菜单示例：\n")
	if len(siblings) == 0 {
		builder.WriteString("（无）\n")
	} else {
		for _, menu := range siblings {
			// 示例路径只展示当前这一级的路径片段
			routePath := strings.TrimPrefix(menu.RoutePath, "/")
			builder.WriteString(fmt.Sprintf(
				"- %s（类型 %s，路径片段 %s，权限 %s）\n",
				menu.Name, menuTypeLabel(menu.Type), blankDefault(routePath, "无"), blankDefault(menu.Perm, "无")))
		}
	}
	return builder.String(), nil
}

// menuTypeLabel 菜单类型取中文名称，无匹配时沿用原值
func menuTypeLabel(menuType string) string {
	switch menuType {
	case "C":
		return "目录"
	case "M":
		return "菜单"
	case "E":
		return "外链"
	case "B":
		return "按钮"
	case "":
		return "未指定"
	default:
		return menuType
	}
}

// blankDefault 值为空时取默认文案
func blankDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// AddMenuForCodegen 代码生成时追加菜单及按钮权限；permPrefix 由 codegen 统一生成，与生成的路由、页面同源
func (s *Service) AddMenuForCodegen(parentMenuId int64, moduleName, businessName, entityName, permPrefix string) error {
	ctx := context.Background()

	parentMenu, err := s.repo.Get(ctx, parentMenuId)
	if err != nil {
		return errs.NotFound("父菜单不存在")
	}

	// 重复保存配置时按路由名去重，避免同一菜单挂出多份按钮
	dup, err := s.repo.RouteNameExists(ctx, entityName, 0)
	if err != nil {
		return errs.SystemError("检查路由名称失败")
	}
	if dup {
		return nil
	}

	entityKebab := utils.KebabCase(entityName)
	menu := &model.Menu{
		ParentID:  types.BigInt(parentMenuId),
		Name:      businessName,
		Type:      "M",
		RouteName: entityName,
		RoutePath: entityKebab,
		Component: moduleName + "/" + entityKebab + "/index",
		Sort:      s.resolveNextSort(ctx, parentMenuId),
		Visible:   1,
		TreePath:  fmt.Sprintf("%s,%d", parentMenu.TreePath, parentMenuId),
	}
	if err := s.repo.Create(ctx, menu); err != nil {
		return errs.SystemError("创建菜单失败")
	}

	// 按钮沿用菜单管理的既有实现，保证字段与手工新建菜单一致
	s.saveCrudButtons(ctx, int64(menu.ID), permPrefix, menu.TreePath)

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

// AddMenuForCodegen 包级桥接（codegen 模块未迁移，走全局 DB 装配）
func AddMenuForCodegen(parentMenuId int64, moduleName, businessName, entityName, permPrefix string) error {
	return NewService(repository.NewRepository(database.DB), nil).AddMenuForCodegen(parentMenuId, moduleName, businessName, entityName, permPrefix)
}
