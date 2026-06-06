package repository

import (
	"strings"

	"gorm.io/gorm"

	"youlai-gin/internal/system/menu/model"
)

// Repository 菜单数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// GetMenuList 菜单列表查询
func (r *Repository) GetMenuList(query *model.MenuQuery) ([]model.Menu, error) {
	var menus []model.Menu
	db := r.db.Model(&model.Menu{})
	if query.Keywords != "" { db = db.Where("name LIKE ?", "%"+query.Keywords+"%") }
	if query.Status != nil { db = db.Where("visible = ?", *query.Status) }
	err := db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// GetMenuByID 根据ID查询菜单
func (r *Repository) GetMenuByID(id int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	return &menu, err
}

// CreateMenu 创建菜单
func (r *Repository) CreateMenu(menu *model.Menu) error { return r.db.Create(menu).Error }

// UpdateMenu 更新菜单
func (r *Repository) UpdateMenu(menu *model.Menu) error { return r.db.Model(&model.Menu{}).Where("id = ?", menu.ID).Updates(menu).Error }

// DeleteMenu 删除菜单
func (r *Repository) DeleteMenu(id int64) error { return r.db.Delete(&model.Menu{}, id).Error }

// GetMenuOptions 获取菜单选项
func (r *Repository) GetMenuOptions(onlyParent bool) ([]model.Menu, error) {
	var menus []model.Menu
	db := r.db.Model(&model.Menu{}).Where("visible = 1")
	if onlyParent { db = db.Where("type IN ('C','M')") }
	err := db.Order("sort ASC").Find(&menus).Error
	return menus, err
}

// GetUserMenus 获取用户菜单（用于路由生成）
func (r *Repository) GetUserMenus(userId int64) ([]model.Menu, error) {
	var menus []model.Menu
	var roleCodes []string
	r.db.Table("sys_user_role ur").Select("r.code").
		Joins("INNER JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 1 AND r.is_deleted = 0", userId).
		Pluck("r.code", &roleCodes)
	var isROOT bool
	for _, c := range roleCodes { if c == "ROOT" { isROOT = true; break } }
	if isROOT {
		err := r.db.Raw("SELECT DISTINCT m.* FROM sys_menu m WHERE m.type IN ('C','M') ORDER BY m.sort ASC, m.id ASC").Scan(&menus).Error
		return menus, err
	}
	err := r.db.Raw(`SELECT DISTINCT m.* FROM sys_menu m INNER JOIN sys_role_menu rm ON m.id = rm.menu_id INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id INNER JOIN sys_role r ON ur.role_id = r.id WHERE ur.user_id = ? AND r.status = 1 AND m.type IN ('C','M') ORDER BY m.sort ASC, m.id ASC`, userId).Scan(&menus).Error
	return menus, err
}

// GetUserButtonPerms 获取用户按钮权限标识列表
func (r *Repository) GetUserButtonPerms(userId int64) ([]string, error) {
	perms := make([]string, 0)
	var isAdmin int
	r.db.Raw("SELECT COUNT(DISTINCT r.id) FROM sys_role r INNER JOIN sys_user_role ur ON r.id = ur.role_id WHERE ur.user_id = ? AND r.code = 'ROOT' AND r.status = 1", userId).Scan(&isAdmin)
	if isAdmin > 0 {
		rows := make([]struct{ Perm string }, 0)
		err := r.db.Raw("SELECT DISTINCT m.perm FROM sys_menu m WHERE m.visible = 1 AND m.type = 'B' AND m.perm IS NOT NULL AND m.perm != ''").Scan(&rows).Error
		if err != nil { return nil, err }
		for _, r := range rows { if p := strings.TrimSpace(r.Perm); p != "" { perms = append(perms, p) } }
		return perms, nil
	}
	rows := make([]struct{ Perm string }, 0)
	err := r.db.Raw(`SELECT DISTINCT m.perm FROM sys_menu m INNER JOIN sys_role_menu rm ON m.id = rm.menu_id INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id INNER JOIN sys_role r ON ur.role_id = r.id WHERE ur.user_id = ? AND r.status = 1 AND m.visible = 1 AND m.type = 'B' AND m.perm IS NOT NULL AND m.perm != ''`, userId).Scan(&rows).Error
	if err != nil { return nil, err }
	for _, r := range rows { if p := strings.TrimSpace(r.Perm); p != "" { perms = append(perms, p) } }
	return perms, nil
}

// CheckMenuNameExists 检查同级菜单名称是否存在
func (r *Repository) CheckMenuNameExists(name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := r.db.Model(&model.Menu{}).Where("name = ? AND parent_id = ?", name, parentId)
	if excludeId > 0 { db = db.Where("id != ?", excludeId) }
	err := db.Count(&count).Error
	return count > 0, err
}

// GetChildrenCount 获取子菜单数量
func (r *Repository) GetChildrenCount(parentId int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Menu{}).Where("parent_id = ?", parentId).Count(&count).Error
	return count, err
}

// GetMaxSortMenuByParentID 获取父菜单下最大排序的菜单
func (r *Repository) GetMaxSortMenuByParentID(parentId int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.Where("parent_id = ?", parentId).Order("sort DESC").First(&menu).Error
	if err != nil { return nil, err }
	return &menu, nil
}
