package repository

import (
	"youlai-gin/internal/database"
)

// RolePerms 角色权限结构
type RolePerms struct {
	RoleCode string   `json:"roleCode"`
	Perms    []string `json:"perms"`
}

// GetAllRolePerms 获取所有角色的权限列表（用于初始化缓存）
func GetAllRolePerms() ([]RolePerms, error) {
	return getRolePermsByCondition("")
}

// GetRolePermsByCode 获取指定角色的权限列表
func GetRolePermsByCode(roleCode string) (*RolePerms, error) {
	rolePermsList, err := getRolePermsByCondition(roleCode)
	if err != nil {
		return nil, err
	}
	
	if len(rolePermsList) == 0 {
		return &RolePerms{
			RoleCode: roleCode,
			Perms:    []string{},
		}, nil
	}
	
	return &rolePermsList[0], nil
}

// GetRolePermsByCodes 批量获取角色权限（用于降级查询）
func GetRolePermsByCodes(roleCodes []string) ([]RolePerms, error) {
	if len(roleCodes) == 0 {
		return []RolePerms{}, nil
	}
	
	// 查询角色和对应的权限按钮
	// type = 4 表示按钮（MenuTypeEnum.BUTTON: 1-菜单 2-目录 3-外链 4-按钮）
	var results []struct {
		RoleCode string
		Perm     string
	}
	
	err := database.DB.Table("sys_role_menu t1").
		Select("t2.code as role_code, t3.perm").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Joins("INNER JOIN sys_menu t3 ON t1.menu_id = t3.id").
		Where("t2.code IN ? AND t3.type = 4 AND t3.perm IS NOT NULL AND t3.perm != ''", roleCodes).
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	// 按角色编码分组权限
	rolePermsMap := make(map[string][]string)
	for _, result := range results {
		rolePermsMap[result.RoleCode] = append(rolePermsMap[result.RoleCode], result.Perm)
	}
	
	// 转换为数组（保持输入顺序）
	rolePermsList := make([]RolePerms, 0, len(roleCodes))
	for _, roleCode := range roleCodes {
		rolePermsList = append(rolePermsList, RolePerms{
			RoleCode: roleCode,
			Perms:    rolePermsMap[roleCode],
		})
	}
	
	return rolePermsList, nil
}

// getRolePermsByCondition 内部方法：根据条件查询角色权限
func getRolePermsByCondition(roleCode string) ([]RolePerms, error) {
	// 查询角色和对应的权限按钮
	// type = 4 表示按钮（MenuTypeEnum.BUTTON: 1-菜单 2-目录 3-外链 4-按钮）
	var results []struct {
		RoleCode string
		Perm     string
	}
	
	query := database.DB.Table("sys_role_menu t1").
		Select("t2.code as role_code, t3.perm").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Joins("INNER JOIN sys_menu t3 ON t1.menu_id = t3.id").
		Where("t3.type = 4 AND t3.perm IS NOT NULL AND t3.perm != ''")
	
	// 如果指定角色编码，添加过滤条件
	if roleCode != "" {
		query = query.Where("t2.code = ?", roleCode)
	}
	
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}
	
	// 按角色编码分组权限
	rolePermsMap := make(map[string][]string)
	for _, result := range results {
		rolePermsMap[result.RoleCode] = append(rolePermsMap[result.RoleCode], result.Perm)
	}
	
	// 转换为数组
	rolePermsList := make([]RolePerms, 0, len(rolePermsMap))
	for code, perms := range rolePermsMap {
		rolePermsList = append(rolePermsList, RolePerms{
			RoleCode: code,
			Perms:    perms,
		})
	}
	
	return rolePermsList, nil
}

// GetRolesAffectedByMenus 获取受菜单影响的角色编码列表（用于菜单变更时刷新缓存）
func GetRolesAffectedByMenus(menuIds []int64) ([]string, error) {
	if len(menuIds) == 0 {
		return []string{}, nil
	}
	
	var roleCodes []string
	err := database.DB.Table("sys_role_menu t1").
		Select("DISTINCT t2.code").
		Joins("INNER JOIN sys_role t2 ON t1.role_id = t2.id AND t2.is_deleted = 0 AND t2.status = 1").
		Where("t1.menu_id IN ?", menuIds).
		Pluck("t2.code", &roleCodes).Error
	
	return roleCodes, err
}
