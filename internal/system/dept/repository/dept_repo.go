package repository

import (
	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/dept/model"
)

// GetDeptList 部门列表查询
func GetDeptList(query *model.DeptQuery) ([]model.Dept, error) {
	var depts []model.Dept
	db := database.DB.Model(&model.Dept{}).Where("is_deleted = 0")

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
func GetDeptByID(id int64) (*model.Dept, error) {
	var dept model.Dept
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&dept).Error
	return &dept, err
}

// CreateDept 创建部门
func CreateDept(dept *model.Dept) error {
	return database.DB.Create(dept).Error
}

// UpdateDept 更新部门
func UpdateDept(dept *model.Dept) error {
	return database.DB.Model(&model.Dept{}).Where("id = ?", dept.ID).Updates(dept).Error
}

// DeleteDept 删除部门（逻辑删除）
func DeleteDept(id int64) error {
	return database.DB.Model(&model.Dept{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetDeptOptions 获取部门下拉选项
func GetDeptOptions() ([]model.Dept, error) {
	var depts []model.Dept
	err := database.DB.Model(&model.Dept{}).
		Where("status = 1 AND is_deleted = 0").
		Order("sort ASC").
		Find(&depts).Error
	return depts, err
}

// CheckDeptNameExists 检查同级部门名称是否存在
func CheckDeptNameExists(name string, parentId int64, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Dept{}).Where("name = ? AND parent_id = ? AND is_deleted = 0", name, parentId)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// CheckDeptCodeExists 检查部门编码是否存在
func CheckDeptCodeExists(code string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Dept{}).Where("code = ? AND is_deleted = 0", code)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetChildrenCount 获取子部门数量
func GetChildrenCount(parentId int64) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Dept{}).Where("parent_id = ? AND is_deleted = 0", parentId).Count(&count).Error
	return count, err
}
