package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	permService "youlai-gin/internal/common/permission/service"
	"youlai-gin/internal/common/redis"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Repository 角色数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.RoleQuery) ([]model.Role, int64, error)
	Get(ctx context.Context, id int64) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, form *model.RoleForm) error
	Delete(ctx context.Context, id int64) error
	Options(ctx context.Context) ([]model.Role, error)
	NameExists(ctx context.Context, name string, excludeID int64) (bool, error)
	CodeExists(ctx context.Context, code string, excludeID int64) (bool, error)
	ListForImport(ctx context.Context) ([]model.Role, error)
	MenuIds(ctx context.Context, roleId int64) ([]int64, error)
	UpdateMenus(ctx context.Context, roleId int64, menuIds []int64) error
	DeptIds(ctx context.Context, roleId int64) ([]int64, error)
	UpdateDepts(ctx context.Context, roleId int64, deptIds []int64) error
	PermsByCode(ctx context.Context, roleCode string) (*model.RolePerms, error)
	PermsByCodes(ctx context.Context, roleCodes []string) ([]model.RolePerms, error)
	CodesByMenuIds(ctx context.Context, menuIds []int64) ([]string, error)
}

// UserIDLister 用户ID查询子集（失效角色关联用户登录态，由 user.Repository 实现）
type UserIDLister interface {
	ListIDsByRoleID(ctx context.Context, roleId int64) ([]int64, error)
}

// Service 角色业务逻辑层
type Service struct {
	repo    Repository
	userIDs UserIDLister
}

// NewService 创建 Service 实例
func NewService(repo Repository, userIDs UserIDLister) *Service {
	return &Service{repo: repo, userIDs: userIDs}
}

// Page 角色分页列表
func (s *Service) Page(ctx context.Context, query *model.RoleQuery) (*baseModel.PagedData, error) {
	roles, total, err := s.repo.Page(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询角色列表失败")
	}

	voList := make([]model.RolePageVO, len(roles))
	for i, role := range roles {
		voList[i] = model.RolePageVO{
			ID:             role.ID,
			Name:           role.Name,
			Code:           role.Code,
			Sort:           role.Sort,
			Status:         role.Status,
			DataScope:      role.DataScope,
			DataScopeLabel: getDataScopeLabel(role.DataScope),
			CreateTime:     types.LocalTime(role.CreateTime),
			UpdateTime:     types.LocalTime(role.UpdateTime),
		}
	}

	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// Options 角色下拉选项
func (s *Service) Options(ctx context.Context) ([]baseModel.Option[types.BigInt], error) {
	roles, err := s.repo.Options(ctx)
	if err != nil {
		return nil, errs.SystemError("查询角色选项失败")
	}

	options := make([]baseModel.Option[types.BigInt], len(roles))
	for i, role := range roles {
		options[i] = baseModel.Option[types.BigInt]{
			Value: role.ID,
			Label: role.Name,
		}
	}

	return options, nil
}

// Create 新增角色
func (s *Service) Create(ctx context.Context, form *model.RoleForm) error {
	if err := s.checkUnique(ctx, form.Name, form.Code, 0); err != nil {
		return err
	}

	role := &model.Role{
		Name:      form.Name,
		Code:      form.Code,
		Sort:      form.Sort,
		Status:    form.Status,
		DataScope: form.DataScope,
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return errs.SystemError("创建角色失败")
	}
	form.ID = role.ID

	return s.applyGrants(ctx, form, 0)
}

// Update 更新角色
// 数据权限范围变更时，失效该角色关联用户的登录态。
func (s *Service) Update(ctx context.Context, id int64, form *model.RoleForm) error {
	form.ID = types.BigInt(id)

	var oldDataScope int
	if oldRole, err := s.repo.Get(ctx, id); err == nil && oldRole != nil {
		oldDataScope = oldRole.DataScope
	}

	if err := s.checkUnique(ctx, form.Name, form.Code, id); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新角色失败")
	}

	return s.applyGrants(ctx, form, oldDataScope)
}

// checkUnique 校验角色名称/编码唯一性
func (s *Service) checkUnique(ctx context.Context, name, code string, excludeID int64) error {
	exists, err := s.repo.NameExists(ctx, name, excludeID)
	if err != nil {
		return errs.SystemError("检查角色名称失败")
	}
	if exists {
		return errs.Business("角色名称已存在")
	}

	exists, err = s.repo.CodeExists(ctx, code, excludeID)
	if err != nil {
		return errs.SystemError("检查角色编码失败")
	}
	if exists {
		return errs.Business("角色编码已存在")
	}

	return nil
}

// applyGrants 保存角色菜单与自定义部门授权
// 数据权限范围变更时，失效该角色关联用户的登录态。
func (s *Service) applyGrants(ctx context.Context, form *model.RoleForm, oldDataScope int) error {
	roleId := int64(form.ID)

	if oldDataScope != 0 && oldDataScope != form.DataScope {
		s.invalidateRoleUsersSessions(roleId)
	}

	if len(form.MenuIds) > 0 {
		menuIds := make([]int64, len(form.MenuIds))
		for i, id := range form.MenuIds {
			menuIds[i] = int64(id)
		}
		if err := s.repo.UpdateMenus(ctx, roleId, menuIds); err != nil {
			return errs.SystemError("更新角色菜单失败")
		}
	}

	// 自定义数据权限：同步 sys_role_dept；非自定义时清理历史自定义部门
	if form.DataScope == permService.DataScopeCustom {
		deptIds := make([]int64, 0, len(form.DeptIds))
		for _, id := range form.DeptIds {
			if int64(id) > 0 {
				deptIds = append(deptIds, int64(id))
			}
		}
		return s.AssignDepts(ctx, roleId, deptIds)
	}

	if err := s.repo.UpdateDepts(ctx, roleId, nil); err != nil {
		return errs.SystemError("清理角色自定义部门失败")
	}

	return nil
}

// GetForm 获取角色表单数据
func (s *Service) GetForm(ctx context.Context, id int64) (*model.RoleForm, error) {
	role, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("角色不存在")
		}
		return nil, errs.SystemError("查询角色失败")
	}

	menuIds, err := s.repo.MenuIds(ctx, id)
	if err != nil {
		return nil, errs.SystemError("查询角色菜单失败")
	}

	menuIdsBigInt := make([]types.BigInt, len(menuIds))
	for i, menuId := range menuIds {
		menuIdsBigInt[i] = types.BigInt(menuId)
	}

	var deptIdsBigInt []types.BigInt
	if role.DataScope == permService.DataScopeCustom {
		deptIds, err := s.repo.DeptIds(ctx, id)
		if err != nil {
			return nil, errs.SystemError("查询角色自定义部门失败")
		}
		deptIdsBigInt = make([]types.BigInt, len(deptIds))
		for i, deptId := range deptIds {
			deptIdsBigInt[i] = types.BigInt(deptId)
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

// Delete 删除角色
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.Get(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("角色不存在")
		}
		return errs.SystemError("查询角色失败")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errs.SystemError("删除角色失败")
	}

	return nil
}

// MenuIds 获取角色已分配的菜单ID列表
func (s *Service) MenuIds(ctx context.Context, roleId int64) ([]int64, error) {
	menuIds, err := s.repo.MenuIds(ctx, roleId)
	if err != nil {
		return nil, errs.SystemError("查询角色菜单失败")
	}
	return menuIds, nil
}

// AssignMenus 分配角色菜单权限并刷新该角色的权限缓存
func (s *Service) AssignMenus(ctx context.Context, roleId int64, menuIds []int64) error {
	role, err := s.repo.Get(ctx, roleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("角色不存在")
		}
		return errs.SystemError("查询角色失败")
	}

	if err := s.repo.UpdateMenus(ctx, roleId, menuIds); err != nil {
		return errs.SystemError("更新角色菜单失败")
	}

	if err := s.RefreshPermsCacheByCode(role.Code); err != nil {
		return errs.SystemError("刷新角色权限缓存失败")
	}

	return nil
}

// DeptIds 获取角色自定义部门ID列表
func (s *Service) DeptIds(ctx context.Context, roleId int64) ([]int64, error) {
	if roleId <= 0 {
		return []int64{}, nil
	}
	return s.repo.DeptIds(ctx, roleId)
}

// AssignDepts 更新角色自定义部门，并在变更时失效该角色关联用户的登录态
func (s *Service) AssignDepts(ctx context.Context, roleId int64, deptIds []int64) error {
	if roleId <= 0 {
		return errs.BadRequest("无效的角色ID")
	}

	oldDeptIds, _ := s.repo.DeptIds(ctx, roleId)
	if err := s.repo.UpdateDepts(ctx, roleId, deptIds); err != nil {
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
		s.invalidateRoleUsersSessions(roleId)
	}

	return nil
}

// invalidateRoleUsersSessions 失效角色关联用户的登录态（递增 token 版本号）
func (s *Service) invalidateRoleUsersSessions(roleId int64) {
	userIds, err := s.userIDs.ListIDsByRoleID(context.Background(), roleId)
	if err != nil || len(userIds) == 0 {
		return
	}
	for _, userId := range userIds {
		_ = invalidateUserSession(userId)
	}
}

// invalidateUserSession 失效单个用户的登录态
func invalidateUserSession(userId int64) error {
	if userId <= 0 {
		return nil
	}
	key := fmt.Sprintf("%s%d", redis.UserTokenVersion, userId)
	return redis.Client.Incr(context.Background(), key).Err()
}

// getDataScopeLabel 获取数据权限显示名称
func getDataScopeLabel(dataScope int) string {
	labels := map[int]string{
		1: "所有数据",
		2: "部门及子部门数据",
		3: "本部门数据",
		4: "本人数据",
		5: "自定义部门数据",
	}
	if label, ok := labels[dataScope]; ok {
		return label
	}
	return ""
}
