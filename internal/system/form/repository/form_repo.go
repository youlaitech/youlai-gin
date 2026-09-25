package repository

import (
	"context"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/system/form/model"
	menuModel "youlai-gin/internal/system/menu/model"
)

// Repository 动态表单数据访问
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建动态表单仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ── 表单定义 ──────────────────────────────────────────────────

// CreateDefinition 新增表单定义
func (r *Repository) CreateDefinition(ctx context.Context, entity *model.FormDefinition) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// UpdateDefinition 按列名更新表单定义
func (r *Repository) UpdateDefinition(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.FormDefinition{}).Where("id = ?", id).Updates(updates).Error
}

// GetDefinition 按 ID 查询未删除的表单定义
func (r *Repository) GetDefinition(ctx context.Context, id int64) (*model.FormDefinition, error) {
	var entity model.FormDefinition
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&entity).Error
	return &entity, err
}

// GetDefinitionByKey 按 form_key 查询未删除的表单定义
func (r *Repository) GetDefinitionByKey(ctx context.Context, formKey string) (*model.FormDefinition, error) {
	var entity model.FormDefinition
	err := r.db.WithContext(ctx).Where("form_key = ? AND is_deleted = 0", formKey).First(&entity).Error
	return &entity, err
}

// GetPublishedByKey 按 form_key 查询已发布的表单定义
func (r *Repository) GetPublishedByKey(ctx context.Context, formKey string) (*model.FormDefinition, error) {
	var entity model.FormDefinition
	err := r.db.WithContext(ctx).
		Where("form_key = ? AND status = 1 AND is_deleted = 0", formKey).
		First(&entity).Error
	return &entity, err
}

// DefinitionPage 表单定义分页
func (r *Repository) DefinitionPage(ctx context.Context, query *model.FormDefinitionQuery) ([]model.FormDefinition, int64, error) {
	var list []model.FormDefinition
	var total int64

	db := r.db.WithContext(ctx).Model(&model.FormDefinition{}).Where("is_deleted = 0")
	if query.Keywords != "" {
		kw := "%" + query.Keywords + "%"
		db = db.Where("form_name LIKE ? OR form_key LIKE ?", kw, kw)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("id DESC").
		Find(&list).Error
	return list, total, err
}

// DefinitionKeyExists 表单标识是否已存在（excludeID > 0 时排除自身）
func (r *Repository) DefinitionKeyExists(ctx context.Context, formKey string, excludeID int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.FormDefinition{}).
		Where("form_key = ? AND is_deleted = 0", formKey)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SoftDeleteDefinitions 逻辑删除表单定义
func (r *Repository) SoftDeleteDefinitions(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.FormDefinition{}).
		Where("id IN ?", ids).
		Update("is_deleted", 1).Error
}

// ── 表单数据 ──────────────────────────────────────────────────

// CreateData 新增表单数据
func (r *Repository) CreateData(ctx context.Context, entity *model.FormData) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// DataPage 表单数据分页
func (r *Repository) DataPage(ctx context.Context, formID int64, query *model.FormDataQuery) ([]model.FormData, int64, error) {
	var list []model.FormData
	var total int64

	db := r.db.WithContext(ctx).Model(&model.FormData{}).
		Where("form_id = ? AND is_deleted = 0", formID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("id DESC").
		Find(&list).Error
	return list, total, err
}

// GetData 按 ID 查询未删除的表单数据
func (r *Repository) GetData(ctx context.Context, id int64) (*model.FormData, error) {
	var entity model.FormData
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&entity).Error
	return &entity, err
}

// SoftDeleteData 逻辑删除表单数据
func (r *Repository) SoftDeleteData(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.FormData{}).
		Where("id IN ?", ids).
		Update("is_deleted", 1).Error
}

// SoftDeleteDataByFormIDs 按表单定义ID级联逻辑删除数据
func (r *Repository) SoftDeleteDataByFormIDs(ctx context.Context, formIDs []int64) error {
	return r.db.WithContext(ctx).Model(&model.FormData{}).
		Where("form_id IN ? AND is_deleted = 0", formIDs).
		Update("is_deleted", 1).Error
}

// ── 版本快照 ──────────────────────────────────────────────────

// CreateSnapshot 固化版本快照
func (r *Repository) CreateSnapshot(ctx context.Context, entity *model.FormSnapshot) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetSnapshot 按表单与版本查询快照
func (r *Repository) GetSnapshot(ctx context.Context, formID int64, version int) (*model.FormSnapshot, error) {
	var entity model.FormSnapshot
	err := r.db.WithContext(ctx).
		Where("form_id = ? AND version = ? AND is_deleted = 0", formID, version).
		First(&entity).Error
	return &entity, err
}

// SoftDeleteSnapshotsByFormIDs 按表单定义ID级联逻辑删除快照
func (r *Repository) SoftDeleteSnapshotsByFormIDs(ctx context.Context, formIDs []int64) error {
	return r.db.WithContext(ctx).Model(&model.FormSnapshot{}).
		Where("form_id IN ? AND is_deleted = 0", formIDs).
		Update("is_deleted", 1).Error
}

// WorkflowFormList 审批类型且已发布的表单（供下拉选项）
func (r *Repository) WorkflowFormList(ctx context.Context) ([]model.FormDefinition, error) {
	var list []model.FormDefinition
	err := r.db.WithContext(ctx).
		Where("category = ? AND status = 1 AND is_deleted = 0", "workflow").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// NicknamesByUserIDs 按用户ID批量查询昵称
func (r *Repository) NicknamesByUserIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(ids) == 0 {
		return result, nil
	}

	var rows []struct {
		ID       int64  `gorm:"column:id"`
		Nickname string `gorm:"column:nickname"`
	}
	if err := r.db.WithContext(ctx).Table("sys_user").
		Select("id, nickname").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return result, err
	}
	for _, row := range rows {
		result[row.ID] = row.Nickname
	}
	return result, nil
}

// ── 菜单（表单访问菜单生成，直接操作 sys_menu / sys_role_menu）─────

// GetMenu 查询菜单
func (r *Repository) GetMenu(ctx context.Context, id int64) (*menuModel.Menu, error) {
	var entity menuModel.Menu
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	return &entity, err
}

// FindMenuByNameAndType 按名称与类型查询菜单（生成默认目录时判重）
func (r *Repository) FindMenuByNameAndType(ctx context.Context, name, menuType string) (*menuModel.Menu, error) {
	var entity menuModel.Menu
	err := r.db.WithContext(ctx).Where("name = ? AND type = ?", name, menuType).First(&entity).Error
	return &entity, err
}

// CreateMenu 新增菜单
func (r *Repository) CreateMenu(ctx context.Context, entity *menuModel.Menu) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// UpdateMenu 整体保存菜单（走 serializer 写 params 列）
func (r *Repository) UpdateMenu(ctx context.Context, entity *menuModel.Menu) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// DeleteMenuWithRoles 删除菜单及其角色授权
func (r *Repository) DeleteMenuWithRoles(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Exec("DELETE FROM sys_role_menu WHERE menu_id = ?", id).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&menuModel.Menu{}, id).Error
}

// MenuRoleIDs 查询菜单已授权的角色ID
func (r *Repository) MenuRoleIDs(ctx context.Context, menuID int64) ([]int64, error) {
	var roleIDs []int64
	err := r.db.WithContext(ctx).Table("sys_role_menu").
		Where("menu_id = ?", menuID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// AssignMenuRoles 给角色授权菜单（含 tree_path 祖先目录，幂等）
func (r *Repository) AssignMenuRoles(ctx context.Context, menuID int64, treePath string, roleIDs []int64) error {
	grantIDs := map[int64]struct{}{menuID: {}}
	for _, part := range strings.Split(treePath, ",") {
		part = strings.TrimSpace(part)
		if part == "" || part == "0" {
			continue
		}
		if id, err := strconv.ParseInt(part, 10, 64); err == nil {
			grantIDs[id] = struct{}{}
		}
	}

	menuIDs := make([]int64, 0, len(grantIDs))
	for id := range grantIDs {
		menuIDs = append(menuIDs, id)
	}

	var rows []struct {
		RoleID int64 `gorm:"column:role_id"`
		MenuID int64 `gorm:"column:menu_id"`
	}
	if err := r.db.WithContext(ctx).Table("sys_role_menu").
		Select("role_id, menu_id").
		Where("menu_id IN ?", menuIDs).
		Scan(&rows).Error; err != nil {
		return err
	}

	existing := make(map[[2]int64]struct{}, len(rows))
	for _, row := range rows {
		existing[[2]int64{row.RoleID, row.MenuID}] = struct{}{}
	}

	for _, roleID := range roleIDs {
		for _, grantID := range menuIDs {
			if _, ok := existing[[2]int64{roleID, grantID}]; ok {
				continue
			}
			if err := r.db.WithContext(ctx).Exec(
				"INSERT INTO sys_role_menu (role_id, menu_id) VALUES (?, ?)", roleID, grantID,
			).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
