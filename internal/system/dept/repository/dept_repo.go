package repository

import (
	"gorm.io/gorm"

	"youlai-gin/internal/common/permission/datascope"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/common/auth"
)

// Repository 部门数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetDeptList 部门列表查询
func (r *Repository) GetDeptList(query *model.DeptQuery, currentUser *auth.UserDetails) ([]model.Dept, error) {
	var depts []model.Dept
	db := r.db.Model(&model.Dept{}).Where("is_deleted = 0")

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

// GetDeptByID 根据ID查询部门
func (r *Repository) GetDeptByID(id int64) (*model.Dept, error) {
	var dept model.Dept
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&dept).Error
	return &dept, err
}

// CreateDept 创建部门
func (r *Repository) CreateDept(dept *model.Dept) error {
	return r.db.Create(dept).Error
}

// UpdateDept 更新部门
func (r *Repository) UpdateDept(dept *model.Dept) error {
	return r.db.Model(&model.Dept{}).Where("id = ?", dept.ID).Updates(dept).Error
}

// DeleteDept 删除部门（逻辑删除）
func (r *Repository) DeleteDept(id int64) error {
	return r.db.Model(&model.Dept{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetDeptOptions 获取部门下拉选项
func (r *Repository) GetDeptOptions(currentUser *auth.UserDetails) ([]model.Dept, error) {
	var depts []model.Dept
	db := r.db.Model(&model.Dept{}).
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

// CheckDeptNameExists 检查同级部门名称是否存在
func (r *Repository) CheckDeptNameExists(name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := r.db.Model(&model.Dept{}).Where("name = ? AND parent_id = ? AND is_deleted = 0", name, parentId)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// CheckDeptCodeExists 检查部门编码是否存在
func (r *Repository) CheckDeptCodeExists(code string, excludeId int64) (bool, error) {
	var count int64
	db := r.db.Model(&model.Dept{}).Where("code = ? AND is_deleted = 0", code)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetChildrenCount 获取子部门数量
func (r *Repository) GetChildrenCount(parentId int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Dept{}).Where("parent_id = ? AND is_deleted = 0", parentId).Count(&count).Error
	return count, err
}

// GetAllDeptsForImport 获取所有部门（用于导入时匹配编码或名称）
func (r *Repository) GetAllDeptsForImport() ([]model.Dept, error) {
	var depts []model.Dept
	err := r.db.Model(&model.Dept{}).
		Where("is_deleted = 0").
		Select("id, code, name").
		Find(&depts).Error
	return depts, err
}
