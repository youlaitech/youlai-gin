package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/database"
	"youlai-gin/internal/common/permission/datascope"
	"youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/gormx"
	"youlai-gin/pkg/types"
)

// Repository 用户数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Page 用户分页查询（数据权限按当前用户多角色并集过滤）
func (r *Repository) Page(ctx context.Context, query *model.UserQuery, currentUser *auth.UserDetails) ([]model.UserPageVO, int64, error) {
	var users []model.UserPageVO
	var total int64

	db := r.db.WithContext(ctx).Table("sys_user u").
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
			constant.RoleCodeRoot,
		)

	// 数据权限过滤（多角色并集策略）
	db = db.Scopes(datascope.DataScopeFilter(currentUser, datascope.DataPermissionConfig{
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
		startTime := query.CreateTime[0]
		endTime := query.CreateTime[1]
		// 结束日期拼接 23:59:59，包含当天所有数据
		if len(endTime) == 10 {
			endTime = endTime + " 23:59:59"
		}
		db = db.Where("u.create_time >= ? AND u.create_time <= ?", startTime, endTime)
	}

	db = db.Group("u.id")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Scopes(database.PaginateFromQuery(query)).Order("u.create_time DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Get 根据ID查询用户
func (r *Repository) Get(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&user).Error
	return &user, err
}

// GetByUsername 根据用户名查询用户（用于登录认证）
func (r *Repository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ? AND is_deleted = 0", username).First(&user).Error
	return &user, err
}

// GetByMobile 根据手机号查询用户
func (r *Repository) GetByMobile(ctx context.Context, mobile string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("mobile = ? AND is_deleted = 0", mobile).First(&user).Error
	return &user, err
}

// GetByEmail 根据邮箱查询用户
func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ? AND is_deleted = 0", email).First(&user).Error
	return &user, err
}

// RoleCodes 获取用户启用角色的编码列表
func (r *Repository) RoleCodes(ctx context.Context, userID int64) ([]string, error) {
	var roleCodes []string
	err := r.db.WithContext(ctx).Table("sys_user_role ur").
		Select("r.code").
		Joins("INNER JOIN sys_role r ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.is_deleted = 0 AND r.status = 1", userID).
		Pluck("r.code", &roleCodes).Error
	return roleCodes, err
}

// Create 创建用户（ctx 携带操作人，由审计钩子填充 create_by/update_by）
func (r *Repository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
// 用 BuildPatchMap(form) 生成「列名→值」映射，从根上解决 GORM Updates(struct) 默认跳过零值字段的问题。
func (r *Repository) Update(ctx context.Context, form *model.UserForm) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", form.ID).
		Updates(gormx.BuildPatchMap(form)).Error
}

// BatchDelete 批量逻辑删除用户
func (r *Repository) BatchDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id IN ?", ids).Update("is_deleted", 1).Error
}

// UpdateStatus 更新用户状态
func (r *Repository) UpdateStatus(ctx context.Context, userId int64, status int) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("status", status).Error
}

// UsernameExists 检查用户名是否存在（排除指定ID）
func (r *Repository) UsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ? AND is_deleted = 0", username)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// RoleIDs 获取用户已分配的角色ID列表
func (r *Repository) RoleIDs(ctx context.Context, userId int64) ([]int64, error) {
	var roleIds []int64
	err := r.db.WithContext(ctx).Model(&model.UserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	return roleIds, err
}

// ListIDsByRoleID 获取角色绑定的用户ID集合
func (r *Repository) ListIDsByRoleID(ctx context.Context, roleId int64) ([]int64, error) {
	var userIds []int64
	err := r.db.WithContext(ctx).Table("sys_user_role").
		Select("user_id").
		Where("role_id = ?", roleId).
		Distinct().
		Pluck("user_id", &userIds).Error
	return userIds, err
}

// UpdateRoles 更新用户角色关联（事务：先删后增）
func (r *Repository) UpdateRoles(ctx context.Context, userId int64, roleIds []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}

		if len(roleIds) == 0 {
			return nil
		}

		userRoles := make([]model.UserRole, len(roleIds))
		for i, roleId := range roleIds {
			userRoles[i] = model.UserRole{
				UserID: types.BigInt(userId),
				RoleID: types.BigInt(roleId),
			}
		}
		return tx.Create(&userRoles).Error
	})
}

// Profile 获取个人中心用户信息（含部门与角色名）
func (r *Repository) Profile(ctx context.Context, userId int64) (*model.UserProfileVO, error) {
	var profile model.UserProfileVO
	err := r.db.WithContext(ctx).Table("sys_user u").
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

// UpdateProfile 更新个人中心信息（仅非空字段）
func (r *Repository) UpdateProfile(ctx context.Context, userId int64, form *model.UserProfileForm) error {
	updates := map[string]interface{}{}
	if form.Nickname != "" {
		updates["nickname"] = form.Nickname
	}
	if form.Avatar != "" {
		updates["avatar"] = form.Avatar
	}
	if form.Gender != nil {
		updates["gender"] = *form.Gender
	}
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Updates(updates).Error
}

// UpdatePassword 更新用户密码
func (r *Repository) UpdatePassword(ctx context.Context, userId int64, password string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("password", password).Error
}

// UpdateMobile 更新用户手机号
func (r *Repository) UpdateMobile(ctx context.Context, userId int64, mobile string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("mobile", mobile).Error
}

// UnbindMobile 解绑用户手机号
func (r *Repository) UnbindMobile(ctx context.Context, userId int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("mobile", nil).Error
}

// UpdateEmail 更新用户邮箱
func (r *Repository) UpdateEmail(ctx context.Context, userId int64, email string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("email", email).Error
}

// UnbindEmail 解绑用户邮箱
func (r *Repository) UnbindEmail(ctx context.Context, userId int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userId).Update("email", nil).Error
}

// Options 获取启用状态的用户下拉选项
func (r *Repository) Options(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("id, username, nickname").
		Where("status = 1 AND is_deleted = 0").
		Order("id ASC").
		Find(&users).Error
	return users, err
}

// GetSocial 查询第三方账号绑定（platform + openid 唯一）
func (r *Repository) GetSocial(ctx context.Context, platform model.SocialPlatform, openID string) (*model.UserSocial, error) {
	var social model.UserSocial
	err := r.db.WithContext(ctx).Where("platform = ? AND openid = ?", platform, openID).First(&social).Error
	return &social, err
}

// UpsertSocial 保存第三方账号绑定：已存在则更新绑定的用户与会话密钥，否则新增
func (r *Repository) UpsertSocial(ctx context.Context, social *model.UserSocial) error {
	err := r.db.WithContext(ctx).
		Where("platform = ? AND openid = ?", social.Platform, social.OpenID).
		Assign(social).
		FirstOrCreate(social).Error
	return err
}

// CreateUserWithGuestRole 事务创建用户并分配游客(GUEST)角色
func (r *Repository) CreateUserWithGuestRole(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserRole{
			UserID: user.ID,
			RoleID: types.BigInt(constant.RoleGuestID),
		}).Error
	})
}
