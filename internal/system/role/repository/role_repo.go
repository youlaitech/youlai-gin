package repository

import (
	"gorm.io/gorm"

	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/pkg/types"
)

// GetRolePage 角色分页查询
func GetRolePage(query *model.RoleQuery) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	db := database.DB.Model(&model.Role{}).Where("is_deleted = 0").Where("code <> ?", "ROOT")

	if query.Keywords != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := query.GetOffset()
	limit := query.GetLimit()
	if err := db.Offset(offset).Limit(limit).Order("sort ASC, create_time DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetRoleByID 根据ID查询角色
func GetRoleByID(id int64) (*model.Role, error) {
	var role model.Role
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&role).Error
	return &role, err
}

// CreateRole 创建角色
func CreateRole(role *model.Role) error {
	return database.DB.Create(role).Error
}

// UpdateRole 更新角色
func UpdateRole(role *model.Role) error {
	return database.DB.Model(&model.Role{}).Where("id = ?", role.ID).Updates(role).Error
}

// DeleteRole 删除角色（逻辑删除）
func DeleteRole(id int64) error {
	return database.DB.Model(&model.Role{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetRoleOptions 获取角色下拉选项
func GetRoleOptions() ([]model.Role, error) {
	var roles []model.Role
	err := database.DB.Model(&model.Role{}).
		Where("status = 1 AND is_deleted = 0").
		Where("code <> ?", "ROOT").
		Order("sort ASC").
		Find(&roles).Error
	return roles, err
}

// GetRoleMenuIds 获取角色已分配的菜单ID列表
func GetRoleMenuIds(roleId int64) ([]int64, error) {
	var menuIds []int64
	err := database.DB.Model(&model.RoleMenu{}).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &menuIds).Error
	return menuIds, err
}

// UpdateRoleMenus 更新角色菜单权限
func UpdateRoleMenus(roleId int64, menuIds []int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
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

// GetRoleDeptIds 获取角色已分配的自定义部门ID列表
func GetRoleDeptIds(roleId int64) ([]int64, error) {
	var deptIds []int64
	err := database.DB.Table("sys_role_dept").
		Where("role_id = ?", roleId).
		Pluck("dept_id", &deptIds).Error
	return deptIds, err
}

// UpdateRoleDepts 更新角色自定义部门（先删后增）
func UpdateRoleDepts(roleId int64, deptIds []int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
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

// CheckRoleNameExists 检查角色名称是否存在
func CheckRoleNameExists(name string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Role{}).Where("name = ? AND is_deleted = 0", name)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// CheckRoleCodeExists 检查角色编码是否存在
func CheckRoleCodeExists(code string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Role{}).Where("code = ? AND is_deleted = 0", code)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}
