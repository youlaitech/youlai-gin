package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/gormx"
	"youlai-gin/pkg/types"
)

// Repository 角色数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Page 角色分页查询
func (r *Repository) Page(ctx context.Context, query *model.RoleQuery) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Role{}).
		Where("is_deleted = 0").
		Where("code <> ?", constant.RoleCodeRoot)

	if query.Keywords != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("sort ASC, create_time DESC").
		Find(&roles).Error

	return roles, total, err
}

// Get 根据ID查询角色
func (r *Repository) Get(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&role).Error
	return &role, err
}

// Create 创建角色（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update 更新角色
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入，
// 从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) Update(ctx context.Context, form *model.RoleForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Role{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// Delete 删除角色（逻辑删除）
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Role{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// Options 获取角色下拉选项
func (r *Repository) Options(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Model(&model.Role{}).
		Where("status = 1 AND is_deleted = 0").
		Where("code <> ?", constant.RoleCodeRoot).
		Order("sort ASC").
		Find(&roles).Error
	return roles, err
}

// NameExists 检查角色名称是否存在（排除指定ID）
func (r *Repository) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Role{}).Where("name = ? AND is_deleted = 0", name)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// CodeExists 检查角色编码是否存在（排除指定ID）
func (r *Repository) CodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Role{}).Where("code = ? AND is_deleted = 0", code)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// ListForImport 获取所有角色（用于导入时匹配编码或名称）
func (r *Repository) ListForImport(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Model(&model.Role{}).
		Where("status = 1 AND is_deleted = 0").
		Select("id, code, name").
		Find(&roles).Error
	return roles, err
}

// MenuIds 获取角色已分配的菜单ID列表
func (r *Repository) MenuIds(ctx context.Context, roleId int64) ([]int64, error) {
	var menuIds []int64
	err := r.db.WithContext(ctx).Model(&model.RoleMenu{}).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &menuIds).Error
	return menuIds, err
}

// UpdateMenus 更新角色菜单权限（事务：先删后增）
func (r *Repository) UpdateMenus(ctx context.Context, roleId int64, menuIds []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleId).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}

		if len(menuIds) > 0 {
			roleMenus := make([]model.RoleMenu, len(menuIds))
			for i, menuId := range menuIds {
				roleMenus[i] = model.RoleMenu{
					RoleID: types.BigInt(roleId),
					MenuID: types.BigInt(menuId),
				}
			}
			if err := tx.Create(&roleMenus).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeptIds 获取角色已分配的自定义部门ID列表
func (r *Repository) DeptIds(ctx context.Context, roleId int64) ([]int64, error) {
	var deptIds []int64
	err := r.db.WithContext(ctx).Table("sys_role_dept").
		Where("role_id = ?", roleId).
		Pluck("dept_id", &deptIds).Error
	return deptIds, err
}

// UpdateDepts 更新角色自定义部门（先删后增）
func (r *Repository) UpdateDepts(ctx context.Context, roleId int64, deptIds []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_dept WHERE role_id = ?", roleId).Error; err != nil {
			return err
		}

		if len(deptIds) == 0 {
			return nil
		}

		rows := make([]map[string]interface{}, 0, len(deptIds))
		seen := make(map[int64]struct{}, len(deptIds))
		for _, deptId := range deptIds {
			if deptId <= 0 {
				continue
			}
			if _, ok := seen[deptId]; ok {
				continue
			}
			seen[deptId] = struct{}{}
			rows = append(rows, map[string]interface{}{"role_id": roleId, "dept_id": deptId})
		}

		if len(rows) == 0 {
			return nil
		}

		return tx.Table("sys_role_dept").Create(&rows).Error
	})
}

// PermsByCode 获取指定角色的权限集合
func (r *Repository) PermsByCode(ctx context.Context, roleCode string) (*model.RolePerms, error) {
	rolePermsList, err := r.permsByCondition(ctx, roleCode)
	if err != nil {
		return nil, err
	}

	if len(rolePermsList) == 0 {
		return &model.RolePerms{
			RoleCode: roleCode,
			Perms:    []string{},
		}, nil
	}

	return &rolePermsList[0], nil
}

// PermsByCodes 批量获取角色权限（用于降级查询，保持输入顺序）
func (r *Repository) PermsByCodes(ctx context.Context, roleCodes []string) ([]model.RolePerms, error) {
	if len(roleCodes) == 0 {
		return []model.RolePerms{}, nil
	}

	var results []struct {
		RoleCode string
		Perm     string
	}

	err := r.db.WithContext(ctx).Table("sys_role_menu t1").
		Select("t2.code as role_code, t3.perm").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Joins("INNER JOIN sys_menu t3 ON t1.menu_id = t3.id").
		Where("t2.code IN ? AND t3.type = 'B' AND t3.perm IS NOT NULL AND t3.perm != ''", roleCodes).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	rolePermsMap := make(map[string][]string)
	for _, result := range results {
		rolePermsMap[result.RoleCode] = append(rolePermsMap[result.RoleCode], result.Perm)
	}

	rolePermsList := make([]model.RolePerms, 0, len(roleCodes))
	for _, roleCode := range roleCodes {
		rolePermsList = append(rolePermsList, model.RolePerms{
			RoleCode: roleCode,
			Perms:    rolePermsMap[roleCode],
		})
	}

	return rolePermsList, nil
}

// CodesByMenuIds 获取受菜单影响的角色编码列表（用于菜单变更时刷新缓存）
func (r *Repository) CodesByMenuIds(ctx context.Context, menuIds []int64) ([]string, error) {
	if len(menuIds) == 0 {
		return []string{}, nil
	}

	var roleCodes []string
	err := r.db.WithContext(ctx).Table("sys_role_menu t1").
		Select("DISTINCT t2.code").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Where("t1.menu_id IN ?", menuIds).
		Pluck("t2.code", &roleCodes).Error

	return roleCodes, err
}

// permsByCondition 根据角色编码查询权限集合（编码为空时查询全部）
func (r *Repository) permsByCondition(ctx context.Context, roleCode string) ([]model.RolePerms, error) {
	// type = 'B' 表示按钮
	var results []struct {
		RoleCode string
		Perm     string
	}

	query := r.db.WithContext(ctx).Table("sys_role_menu t1").
		Select("t2.code as role_code, t3.perm").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Joins("INNER JOIN sys_menu t3 ON t1.menu_id = t3.id").
		Where("t3.type = 'B' AND t3.perm IS NOT NULL AND t3.perm != ''")

	if roleCode != "" {
		query = query.Where("t2.code = ?", roleCode)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	rolePermsMap := make(map[string][]string)
	for _, result := range results {
		rolePermsMap[result.RoleCode] = append(rolePermsMap[result.RoleCode], result.Perm)
	}

	rolePermsList := make([]model.RolePerms, 0, len(rolePermsMap))
	for code, perms := range rolePermsMap {
		rolePermsList = append(rolePermsList, model.RolePerms{
			RoleCode: code,
			Perms:    perms,
		})
	}

	return rolePermsList, nil
}
