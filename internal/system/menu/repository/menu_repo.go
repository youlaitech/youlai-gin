package repository

import (
	"strings"

	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/menu/model"
)

// GetMenuList 菜单列表查询
func GetMenuList(query *model.MenuQuery) ([]model.Menu, error) {
	var menus []model.Menu
	db := database.DB.Model(&model.Menu{})

	if query.Keywords != "" {
		db = db.Where("name LIKE ?", "%"+query.Keywords+"%")
	}

	if query.Status != nil {
		db = db.Where("visible = ?", *query.Status)
	}

	err := db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// GetMenuByID 根据ID查询菜单
func GetMenuByID(id int64) (*model.Menu, error) {
	var menu model.Menu
	err := database.DB.Where("id = ?", id).First(&menu).Error
	return &menu, err
}

// CreateMenu 创建菜单
func CreateMenu(menu *model.Menu) error {
	return database.DB.Create(menu).Error
}

// UpdateMenu 更新菜单
func UpdateMenu(menu *model.Menu) error {
	return database.DB.Model(&model.Menu{}).Where("id = ?", menu.ID).Updates(menu).Error
}

// DeleteMenu 删除菜单
func DeleteMenu(id int64) error {
	return database.DB.Delete(&model.Menu{}, id).Error
}

// GetMenuOptions 获取菜单选项
func GetMenuOptions(onlyParent bool) ([]model.Menu, error) {
	var menus []model.Menu
	db := database.DB.Model(&model.Menu{}).Where("visible = 1")

	if onlyParent {
		db = db.Where("type IN ('C','M')")
	}

	err := db.Order("sort ASC").Find(&menus).Error
	return menus, err
}

// GetUserMenus 获取用户菜单（用于路由生成）
func GetUserMenus(userId int64) ([]model.Menu, error) {
	var menus []model.Menu
	
	// 查询用户是否是超级管理员（ROOT）
	var isAdmin int
	database.DB.Raw(`
		SELECT COUNT(DISTINCT r.id)
		FROM sys_role r
		INNER JOIN sys_user_role ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.code = 'ROOT' AND r.status = 1
	`, userId).Scan(&isAdmin)
	
	// 超级管理员返回所有菜单
	if isAdmin > 0 {
		err := database.DB.Raw(`
			SELECT DISTINCT m.*
			FROM sys_menu m
			WHERE m.visible = 1
			AND m.type IN ('C','M')
			ORDER BY m.sort ASC, m.id ASC
		`).Scan(&menus).Error
		return menus, err
	}
	
	// 普通用户根据角色权限查询菜单
	err := database.DB.Raw(`
		SELECT DISTINCT m.*
		FROM sys_menu m
		INNER JOIN sys_role_menu rm ON m.id = rm.menu_id
		INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		AND r.status = 1
		AND m.visible = 1
		AND m.type IN ('C','M')
		ORDER BY m.sort ASC, m.id ASC
	`, userId).Scan(&menus).Error
	
	return menus, err
}

// GetUserButtonPerms 获取用户按钮权限标识列表
func GetUserButtonPerms(userId int64) ([]string, error) {
	perms := make([]string, 0)

	// 查询用户是否是超级管理员（ROOT）
	var isAdmin int
	database.DB.Raw(`
		SELECT COUNT(DISTINCT r.id)
		FROM sys_role r
		INNER JOIN sys_user_role ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.code = 'ROOT' AND r.status = 1
	`, userId).Scan(&isAdmin)

	if isAdmin > 0 {
		rows := make([]struct{ Perm string }, 0)
		err := database.DB.Raw(`
			SELECT DISTINCT m.perm
			FROM sys_menu m
			WHERE m.visible = 1
			AND m.type = 'B'
			AND m.perm IS NOT NULL
			AND m.perm != ''
		`).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			p := strings.TrimSpace(r.Perm)
			if p != "" {
				perms = append(perms, p)
			}
		}
		return perms, nil
	}

	rows := make([]struct{ Perm string }, 0)
	err := database.DB.Raw(`
		SELECT DISTINCT m.perm
		FROM sys_menu m
		INNER JOIN sys_role_menu rm ON m.id = rm.menu_id
		INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		AND r.status = 1
		AND m.visible = 1
		AND m.type = 'B'
		AND m.perm IS NOT NULL
		AND m.perm != ''
	`, userId).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		p := strings.TrimSpace(r.Perm)
		if p != "" {
			perms = append(perms, p)
		}
	}
	return perms, nil
}

// CheckMenuNameExists 检查同级菜单名称是否存在
func CheckMenuNameExists(name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Menu{}).Where("name = ? AND parent_id = ?", name, parentId)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetChildrenCount 获取子菜单数量
func GetChildrenCount(parentId int64) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Menu{}).Where("parent_id = ?", parentId).Count(&count).Error
	return count, err
}
