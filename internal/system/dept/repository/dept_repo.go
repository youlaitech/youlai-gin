package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/permission/datascope"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/pkg/gormx"
)

// Repository 部门数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// List 部门列表查询（按数据权限过滤）
func (r *Repository) List(ctx context.Context, query *model.DeptQuery, currentUser *auth.UserDetails) ([]model.Dept, error) {
	var depts []model.Dept
	db := r.db.WithContext(ctx).Model(&model.Dept{}).Where("is_deleted = 0")

	db = db.Scopes(datascope.DataScopeFilter(currentUser, datascope.DataPermissionConfig{
		DeptAlias:    "",
		DeptIDColumn: "id",
		UserAlias:    "",
		UserIDColumn: "create_by",
	}))

	if query.Keywords != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	err := db.Order("sort ASC, id ASC").Find(&depts).Error
	return depts, err
}

// Get 根据ID查询部门
func (r *Repository) Get(ctx context.Context, id int64) (*model.Dept, error) {
	var dept model.Dept
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&dept).Error
	return &dept, err
}

// Create 创建部门（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) Create(ctx context.Context, dept *model.Dept) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

// Update 更新部门
// 用 BuildPatchMap(form) 生成「列名→值」映射：指针字段 nil 跳过、非 nil（含 0）写入，
// 从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) Update(ctx context.Context, form *model.DeptForm) error {
	return r.db.WithContext(ctx).
		Model(&model.Dept{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// Delete 删除部门（逻辑删除）
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Dept{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// Options 部门下拉选项（启用中，按数据权限过滤）
func (r *Repository) Options(ctx context.Context, currentUser *auth.UserDetails) ([]model.Dept, error) {
	var depts []model.Dept
	db := r.db.WithContext(ctx).Model(&model.Dept{}).
		Where("status = 1 AND is_deleted = 0").
		Order("sort ASC")

	db = db.Scopes(datascope.DataScopeFilter(currentUser, datascope.DataPermissionConfig{
		DeptAlias:    "",
		DeptIDColumn: "id",
		UserAlias:    "",
		UserIDColumn: "create_by",
	}))

	err := db.Find(&depts).Error
	return depts, err
}

// NameExists 检查同级部门名称是否存在（excludeId>0 时排除自身）
func (r *Repository) NameExists(ctx context.Context, name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Dept{}).Where("name = ? AND parent_id = ? AND is_deleted = 0", name, parentId)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// CodeExists 检查部门编码是否存在（excludeId>0 时排除自身）
func (r *Repository) CodeExists(ctx context.Context, code string, excludeId int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Dept{}).Where("code = ? AND is_deleted = 0", code)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// ChildrenCount 获取子部门数量
func (r *Repository) ChildrenCount(ctx context.Context, parentId int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Dept{}).Where("parent_id = ? AND is_deleted = 0", parentId).Count(&count).Error
	return count, err
}

// ListForImport 获取所有部门（用于导入时匹配编码或名称）
func (r *Repository) ListForImport(ctx context.Context) ([]model.Dept, error) {
	var depts []model.Dept
	err := r.db.WithContext(ctx).Model(&model.Dept{}).
		Where("is_deleted = 0").
		Select("id, code, name").
		Find(&depts).Error
	return depts, err
}
