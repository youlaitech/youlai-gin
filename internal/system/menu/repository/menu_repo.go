package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/system/menu/model"
	"youlai-gin/pkg/gormx"
)

// Repository 菜单数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// List 菜单列表查询
func (r *Repository) List(ctx context.Context, query *model.MenuQuery) ([]model.Menu, error) {
	var menus []model.Menu
	db := r.db.WithContext(ctx).Model(&model.Menu{})
	if query.Keywords != "" {
		db = db.Where("name LIKE ?", "%"+query.Keywords+"%")
	}
	if query.Status != nil {
		db = db.Where("visible = ?", *query.Status)
	}
	err := db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// Get 根据ID查询菜单
func (r *Repository) Get(ctx context.Context, id int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&menu).Error
	return &menu, err
}

// Create 创建菜单（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) Create(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

// Update 更新菜单
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入，
// 从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) Update(ctx context.Context, form *model.MenuForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Menu{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// Delete 删除菜单（物理删除）
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}

// Options 菜单下拉选项（可见项；onlyParent 时仅目录与菜单类型）
func (r *Repository) Options(ctx context.Context, onlyParent bool) ([]model.Menu, error) {
	var menus []model.Menu
	db := r.db.WithContext(ctx).Model(&model.Menu{}).Where("visible = 1")
	if onlyParent {
		db = db.Where("type IN ('C','M')")
	}
	err := db.Order("sort ASC").Find(&menus).Error
	return menus, err
}

// UserMenus 获取用户菜单（用于路由生成；ROOT 角色返回全量非按钮菜单）
func (r *Repository) UserMenus(ctx context.Context, userId int64) ([]model.Menu, error) {
	var menus []model.Menu
	var roleCodes []string
	r.db.WithContext(ctx).Table("sys_user_role ur").Select("r.code").
		Joins("INNER JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 1 AND r.is_deleted = 0", userId).
		Pluck("r.code", &roleCodes)

	var isROOT bool
	for _, c := range roleCodes {
		if c == "ROOT" {
			isROOT = true
			break
		}
	}

	if isROOT {
		err := r.db.WithContext(ctx).
			Raw("SELECT DISTINCT m.* FROM sys_menu m WHERE m.type != 'B' ORDER BY m.sort ASC, m.id ASC").
			Scan(&menus).Error
		return menus, err
	}

	err := r.db.WithContext(ctx).
		Raw(`SELECT DISTINCT m.* FROM sys_menu m INNER JOIN sys_role_menu rm ON m.id = rm.menu_id INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id INNER JOIN sys_role r ON ur.role_id = r.id WHERE ur.user_id = ? AND r.status = 1 AND m.type != 'B' ORDER BY m.sort ASC, m.id ASC`, userId).
		Scan(&menus).Error
	return menus, err
}

// NameExists 检查同级菜单名称是否存在（excludeId>0 时排除自身）
func (r *Repository) NameExists(ctx context.Context, name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Menu{}).Where("name = ? AND parent_id = ?", name, parentId)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// RouteNameExists 检查路由名称是否存在（excludeId>0 时排除自身）
func (r *Repository) RouteNameExists(ctx context.Context, routeName string, excludeId int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Menu{}).Where("route_name = ?", routeName)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// ChildrenCount 获取子菜单数量
func (r *Repository) ChildrenCount(ctx context.Context, parentId int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Menu{}).Where("parent_id = ?", parentId).Count(&count).Error
	return count, err
}

// MaxSortMenu 获取父菜单下排序最大的菜单（用于代码生成追加排序）
func (r *Repository) MaxSortMenu(ctx context.Context, parentId int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentId).Order("sort DESC").First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// Children 获取直接子菜单
func (r *Repository) Children(ctx context.Context, parentId int64) ([]model.Menu, error) {
	var children []model.Menu
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentId).Find(&children).Error
	return children, err
}

// UpdateTreePathByParent 批量更新子节点树路径
func (r *Repository) UpdateTreePathByParent(ctx context.Context, parentId int64, treePath string) error {
	return r.db.WithContext(ctx).Model(&model.Menu{}).Where("parent_id = ?", parentId).Update("tree_path", treePath).Error
}

// UpdateTreePath 更新单个菜单的树路径（父级变更时调用）
func (r *Repository) UpdateTreePath(ctx context.Context, id int64, treePath string) error {
	return r.db.WithContext(ctx).Model(&model.Menu{}).Where("id = ?", id).Update("tree_path", treePath).Error
}
