package repository

import (
	"gorm.io/gorm"
	
	"youlai-gin/internal/database"
	"youlai-gin/internal/system/role/model"
)

// GetRolePage 角色分页查询
func GetRolePage(query *model.RolePageQuery) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	db := database.DB.Model(&model.Role{}).Where("is_deleted = 0")

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
					RoleID: roleId,
					MenuID: menuId,
				}
			}
			if err := tx.Create(&roleMenus).Error; err != nil {
				return err
			}
		}

		return nil
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

