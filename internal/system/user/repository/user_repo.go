package repository

import (
	"youlai-gin/pkg/database"
	roleRepo "youlai-gin/internal/system/role/repository"
	"youlai-gin/internal/system/user/model"
	pkgDatabase "youlai-gin/pkg/database"
	"youlai-gin/pkg/types"
)

// GetRolePermsByCodes 从数据库查询角色权限（降级用）
func GetRolePermsByCodes(roleCodes []string) ([]roleRepo.RolePerms, error) {
	return roleRepo.GetRolePermsByCodes(roleCodes)
}

// GetUserPage 用户分页查询
func GetUserPage(query *model.UserPageQuery) ([]model.UserPageVO, int64, error) {
	var users []model.UserPageVO
	var total int64

	db := database.DB.Table("sys_user u").
		Select(`u.id, u.username, u.nickname, u.mobile, u.gender, u.avatar, u.email, u.status,
			u.create_time, d.name as dept_name,
			GROUP_CONCAT(r.name ORDER BY r.id SEPARATOR ',') as role_names`).
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
		Joins("LEFT JOIN sys_user_role ur ON u.id = ur.user_id").
		Joins("LEFT JOIN sys_role r ON ur.role_id = r.id").
		Where("u.is_deleted = 0")

	if query.Keywords != "" {
		db = db.Where("u.username LIKE ? OR u.nickname LIKE ? OR u.mobile LIKE ?",
			"%"+query.Keywords+"%", "%"+query.Keywords+"%", "%"+query.Keywords+"%")
	}

	if query.Status != nil {
		db = db.Where("u.status = ?", *query.Status)
	}

	if query.DeptID != nil {
		db = db.Where("u.dept_id = ?", *query.DeptID)
	}

	if len(query.CreateTime) == 2 {
		db = db.Where("u.create_time BETWEEN ? AND ?", query.CreateTime[0], query.CreateTime[1])
	}

	db = db.Group("u.id")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 使用通用分页函数
	if err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).Order("u.create_time DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserByID 根据ID查询用户
func GetUserByID(id int64) (*model.User, error) {
	var user model.User
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&user).Error
	return &user, err
}

// GetUserByUsername 根据用户名查询用户（用于登录认证）
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ? AND is_deleted = 0", username).First(&user).Error
	return &user, err
}

// FindByUsername GetUserByUsername 的别名函数
func FindByUsername(username string) (*model.User, error) {
	return GetUserByUsername(username)
}

// GetUserByMobile 根据手机号查询用户
func GetUserByMobile(mobile string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("mobile = ? AND is_deleted = 0", mobile).First(&user).Error
	return &user, err
}

// GetUserRoles 获取用户角色编码列表
func GetUserRoles(userID int64) ([]string, error) {
	var roleCodes []string
	err := database.DB.Table("sys_user_role ur").
		Select("r.code").
		Joins("INNER JOIN sys_role r ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.is_deleted = 0 AND r.status = 1", userID).
		Pluck("r.code", &roleCodes).Error
	return roleCodes, err
}

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

// UpdateUser 更新用户
func UpdateUser(user *model.User) error {
	return database.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(user).Error
}

// DeleteUser 删除用户（逻辑删除）
func DeleteUser(id int64) error {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// DeleteUsersByIDs 批量删除用户
func DeleteUsersByIDs(ids []int64) error {
	return database.DB.Model(&model.User{}).Where("id IN ?", ids).Update("is_deleted", 1).Error
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(userId int64, status int) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userId).Update("status", status).Error
}

// CheckUsernameExists 检查用户名是否存在
func CheckUsernameExists(username string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&model.User{}).Where("username = ? AND is_deleted = 0", username)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetUserRoleIDs 获取用户角色ID列表
func GetUserRoleIDs(userId int64) ([]int64, error) {
	var roleIds []int64
	err := database.DB.Model(&model.UserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	return roleIds, err
}

// SaveUserRoles 保存用户角色关联（事务：先删除再新增）
func SaveUserRoles(userId int64, roleIds []int64) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除旧的角色关联
	if err := tx.Where("user_id = ?", userId).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 新增角色关联
	if len(roleIds) > 0 {
		userRoles := make([]model.UserRole, len(roleIds))
		for i, roleId := range roleIds {
			userRoles[i] = model.UserRole{
				UserID: types.BigInt(userId),
				RoleID: types.BigInt(roleId),
			}
		}
		if err := tx.Create(&userRoles).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetUserProfile 获取用户个人信息
func GetUserProfile(userId int64) (*model.UserProfileVO, error) {
	var profile model.UserProfileVO
	err := database.DB.Table("sys_user u").
		Select(`u.id, u.username, u.nickname, u.avatar, u.gender, u.mobile, u.email,
			d.name as dept_name,
			GROUP_CONCAT(r.name ORDER BY r.id SEPARATOR ',') as role_names`).
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
		Joins("LEFT JOIN sys_user_role ur ON u.id = ur.user_id").
		Joins("LEFT JOIN sys_role r ON ur.role_id = r.id").
		Where("u.id = ? AND u.is_deleted = 0", userId).
		Group("u.id").
		First(&profile).Error
	return &profile, err
}

// UpdateUserProfile 更新用户个人信息
func UpdateUserProfile(userId int64, form *model.UserProfileForm) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userId).Updates(map[string]interface{}{
		"nickname": form.Nickname,
		"avatar":   form.Avatar,
		"gender":   form.Gender,
		"mobile":   form.Mobile,
		"email":    form.Email,
	}).Error
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userId int64, password string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userId).Update("password", password).Error
}

// UpdateUserMobile 更新用户手机号
func UpdateUserMobile(userId int64, mobile string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userId).Update("mobile", mobile).Error
}

// UpdateUserEmail 更新用户邮箱
func UpdateUserEmail(userId int64, email string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userId).Update("email", email).Error
}

// GetUserOptions 获取用户下拉选项
func GetUserOptions() ([]model.User, error) {
	var users []model.User
	err := database.DB.Model(&model.User{}).
		Select("id, username, nickname").
		Where("status = 1 AND is_deleted = 0").
		Order("id ASC").
		Find(&users).Error
	return users, err
}
