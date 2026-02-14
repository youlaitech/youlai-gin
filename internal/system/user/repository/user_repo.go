package repository

import (
	roleRepo "youlai-gin/internal/system/role/repository"
	"youlai-gin/internal/system/user/api"
	"youlai-gin/internal/system/user/domain"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/database"
	pkgDatabase "youlai-gin/pkg/database"
	"youlai-gin/pkg/middleware"
	"youlai-gin/pkg/types"
)

// GetRolePermsByCodes 从数据库查询角色权限（降级用）
func GetRolePermsByCodes(roleCodes []string) ([]roleRepo.RolePerms, error) {
	return roleRepo.GetRolePermsByCodes(roleCodes)
}

// GetUserPage 用户分页查询
func GetUserPage(query *api.UserQueryReq, currentUser *auth.UserDetails) ([]api.UserPageResp, int64, error) {
	var users []api.UserPageResp
	var total int64

	db := database.DB.Table("sys_user u").
		Select(`u.id, u.username, u.nickname, u.mobile, u.gender, u.avatar, u.email, u.status,
			u.create_time, d.name as dept_name,
			GROUP_CONCAT(r.name ORDER BY r.id SEPARATOR ',') as role_names`).
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
		Joins("LEFT JOIN sys_user_role ur ON u.id = ur.user_id").
		Joins("LEFT JOIN sys_role r ON ur.role_id = r.id").
		Where("u.is_deleted = 0").
		Where(
			`NOT EXISTS (
				SELECT 1
				FROM sys_user_role sur
					INNER JOIN sys_role sr ON sur.role_id = sr.id
				WHERE sur.user_id = u.id
					AND sr.code = ?
			)`,
			"ROOT",
		)

	// 数据权限过滤（多角色并集策略）
	db = db.Scopes(middleware.DataScopeFilter(currentUser, middleware.DataPermissionConfig{
		DeptAlias:    "u",
		DeptIDColumn: "dept_id",
		UserAlias:    "u",
		UserIDColumn: "create_by",
	}))

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
func GetUserByID(id int64) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&user).Error
	return &user, err
}

// GetUserByUsername 根据用户名查询用户（用于登录认证）
func GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("username = ? AND is_deleted = 0", username).First(&user).Error
	return &user, err
}

// FindByUsername GetUserByUsername 的别名函数
func FindByUsername(username string) (*domain.User, error) {
	return GetUserByUsername(username)
}

// GetUserByMobile 根据手机号查询用户
func GetUserByMobile(mobile string) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("mobile = ? AND is_deleted = 0", mobile).First(&user).Error
	return &user, err
}

// GetUserByEmail 根据邮箱查询用户
func GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("email = ? AND is_deleted = 0", email).First(&user).Error
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
func CreateUser(user *domain.User) error {
	return database.DB.Create(user).Error
}

// UpdateUser 更新用户
func UpdateUser(user *domain.User) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", user.ID).Updates(user).Error
}

// DeleteUser 删除用户（逻辑删除）
func DeleteUser(id int64) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// DeleteUsersByIDs 批量删除用户
func DeleteUsersByIDs(ids []int64) error {
	return database.DB.Model(&domain.User{}).Where("id IN ?", ids).Update("is_deleted", 1).Error
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(userId int64, status int) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("status", status).Error
}

// CheckUsernameExists 检查用户名是否存在
func CheckUsernameExists(username string, excludeId int64) (bool, error) {
	var count int64
	db := database.DB.Model(&domain.User{}).Where("username = ? AND is_deleted = 0", username)
	if excludeId > 0 {
		db = db.Where("id != ?", excludeId)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// GetUserRoleIDs 获取用户角色ID列表
func GetUserRoleIDs(userId int64) ([]int64, error) {
	var roleIds []int64
	err := database.DB.Model(&domain.UserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	return roleIds, err
}

// ListUserIDsByRoleID 获取角色绑定的用户ID集合
func ListUserIDsByRoleID(roleId int64) ([]int64, error) {
	var userIds []int64
	err := database.DB.Table("sys_user_role").
		Select("user_id").
		Where("role_id = ?", roleId).
		Distinct().
		Pluck("user_id", &userIds).Error
	return userIds, err
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
	if err := tx.Where("user_id = ?", userId).Delete(&domain.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 新增角色关联
	if len(roleIds) > 0 {
		userRoles := make([]domain.UserRole, len(roleIds))
		for i, roleId := range roleIds {
			userRoles[i] = domain.UserRole{
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
func GetUserProfile(userId int64) (*api.UserProfileResp, error) {
	var profile api.UserProfileResp
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
func UpdateUserProfile(userId int64, req *api.UserProfileUpdateReq) error {
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Updates(updates).Error
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userId int64, password string) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("password", password).Error
}

// UpdateUserMobile 更新用户手机号
func UpdateUserMobile(userId int64, mobile string) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("mobile", mobile).Error
}

// UnbindUserMobile 解绑用户手机号
func UnbindUserMobile(userId int64) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("mobile", nil).Error
}

// UpdateUserEmail 更新用户邮箱
func UpdateUserEmail(userId int64, email string) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("email", email).Error
}

// UnbindUserEmail 解绑用户邮箱
func UnbindUserEmail(userId int64) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", userId).Update("email", nil).Error
}

// GetUserOptions 获取用户下拉选项
func GetUserOptions() ([]domain.User, error) {
	var users []domain.User
	err := database.DB.Model(&domain.User{}).
		Select("id, username, nickname").
		Where("status = 1 AND is_deleted = 0").
		Order("id ASC").
		Find(&users).Error
	return users, err
}
