package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/excel"
	"youlai-gin/internal/common/redis"
	"youlai-gin/internal/common/utils"
	deptModel "youlai-gin/internal/system/dept/model"
	roleModel "youlai-gin/internal/system/role/model"
	"youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Repository 用户数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.UserQuery, currentUser *auth.UserDetails) ([]model.UserPageVO, int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByMobile(ctx context.Context, mobile string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	RoleCodes(ctx context.Context, userID int64) ([]string, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, form *model.UserForm) error
	BatchDelete(ctx context.Context, ids []int64) error
	UpdateStatus(ctx context.Context, userId int64, status int) error
	UsernameExists(ctx context.Context, username string, excludeID int64) (bool, error)
	RoleIDs(ctx context.Context, userId int64) ([]int64, error)
	ListIDsByRoleID(ctx context.Context, roleId int64) ([]int64, error)
	UpdateRoles(ctx context.Context, userId int64, roleIds []int64) error
	Profile(ctx context.Context, userId int64) (*model.UserProfileVO, error)
	UpdateProfile(ctx context.Context, userId int64, form *model.UserProfileForm) error
	UpdatePassword(ctx context.Context, userId int64, password string) error
	UpdateMobile(ctx context.Context, userId int64, mobile string) error
	UnbindMobile(ctx context.Context, userId int64) error
	UpdateEmail(ctx context.Context, userId int64, email string) error
	UnbindEmail(ctx context.Context, userId int64) error
	Options(ctx context.Context) ([]model.User, error)
}

// RoleRepository 角色数据访问子集（导入匹配、权限降级查询）
type RoleRepository interface {
	ListForImport(ctx context.Context) ([]roleModel.Role, error)
	PermsByCodes(ctx context.Context, roleCodes []string) ([]roleModel.RolePerms, error)
}

// RoleCacheRefresher 角色权限缓存刷新（由 role.Service 实现）
type RoleCacheRefresher interface {
	RefreshPermsCacheByCodes(roleCodes []string) error
}

// DeptRepository 部门数据访问子集（导入匹配）
type DeptRepository interface {
	ListForImport(ctx context.Context) ([]deptModel.Dept, error)
}

// Service 用户业务逻辑层
type Service struct {
	repo      Repository
	roleRepo  RoleRepository
	roleCache RoleCacheRefresher
	deptRepo  DeptRepository
}

// NewService 创建 Service 实例
func NewService(repo Repository, roleRepo RoleRepository, roleCache RoleCacheRefresher, deptRepo DeptRepository) *Service {
	return &Service{repo: repo, roleRepo: roleRepo, roleCache: roleCache, deptRepo: deptRepo}
}

// Page 用户分页列表
func (s *Service) Page(ctx context.Context, query *model.UserQuery, currentUser *auth.UserDetails) (*baseModel.PagedData, error) {
	users, total, err := s.repo.Page(ctx, query, currentUser)
	if err != nil {
		return nil, errs.SystemError("查询用户列表失败")
	}

	return &baseModel.PagedData{List: users, Total: total}, nil
}

// Create 新增用户（默认密码为系统配置，可指定角色）
func (s *Service) Create(ctx context.Context, form *model.UserForm) error {
	exists, err := s.repo.UsernameExists(ctx, form.Username, 0)
	if err != nil {
		return errs.SystemError("检查用户名失败")
	}
	if exists {
		return errs.Business("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(constant.DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.SystemError("密码加密失败")
	}

	user := &model.User{
		Username: form.Username,
		Nickname: form.Nickname,
		Mobile:   form.Mobile,
		Gender:   int(form.Gender),
		Email:    form.Email,
		DeptID:   form.DeptID,
		Status:   int(form.Status),
		Avatar:   form.Avatar,
		Password: string(hashedPassword),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return errs.SystemError("创建用户失败").WithErr(err)
	}
	form.ID = user.ID

	if len(form.RoleIDs) > 0 {
		if err := s.repo.UpdateRoles(ctx, int64(user.ID), types.ToInt64Slice(form.RoleIDs)); err != nil {
			return errs.SystemError("分配用户角色失败").WithErr(err)
		}
	}

	return nil
}

// Update 更新用户（角色关联先删后增，空列表即清空）
func (s *Service) Update(ctx context.Context, id int64, form *model.UserForm) error {
	form.ID = types.BigInt(id)

	exists, err := s.repo.UsernameExists(ctx, form.Username, id)
	if err != nil {
		return errs.SystemError("检查用户名失败")
	}
	if exists {
		return errs.Business("用户名已存在")
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return errs.SystemError("更新用户失败").WithErr(err)
	}

	if err := s.repo.UpdateRoles(ctx, id, types.ToInt64Slice(form.RoleIDs)); err != nil {
		return errs.SystemError("更新用户角色失败").WithErr(err)
	}

	return nil
}

// GetForm 获取用户表单数据
func (s *Service) GetForm(ctx context.Context, userId int64) (*model.UserFormVO, error) {
	if userId == 0 {
		return &model.UserFormVO{}, nil
	}

	user, err := s.repo.Get(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	roleIDs, err := s.repo.RoleIDs(ctx, userId)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	return &model.UserFormVO{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Email:    user.Email,
		Avatar:   user.Avatar,
		DeptID:   user.DeptID,
		Status:   user.Status,
		RoleIDs:  types.ToBigIntSlice(roleIDs),
	}, nil
}

// Delete 批量删除用户（逗号分隔ID，非法项跳过）
func (s *Service) Delete(ctx context.Context, ids string) error {
	if ids == "" {
		return errs.BadRequest("请选择要删除的用户")
	}

	userIDs := make([]int64, 0)
	for _, idStr := range strings.Split(ids, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		return errs.BadRequest("无效的用户ID")
	}

	if err := s.repo.BatchDelete(ctx, userIDs); err != nil {
		return errs.SystemError("删除用户失败")
	}

	return nil
}

// UpdateStatus 更新用户状态
func (s *Service) UpdateStatus(ctx context.Context, userId int64, status int) error {
	if err := s.repo.UpdateStatus(ctx, userId, status); err != nil {
		return errs.SystemError("更新用户状态失败")
	}
	return nil
}

// CurrentUser 获取当前登录用户信息（角色取自 token，权限读缓存）
func (s *Service) CurrentUser(ctx context.Context, userId int64, roles []string) (*model.CurrentUserVO, error) {
	user, err := s.repo.Get(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 获取用户权限列表（从Redis缓存）
	perms := []string{}
	if len(roles) > 0 {
		perms, err = s.rolePermsFromCache(ctx, roles)
		if err != nil {
			return nil, errs.SystemError("查询用户权限失败")
		}
	}

	// 缓存未命中时触发全量刷新（防止 DB 数据已更新但缓存未同步）
	if len(perms) == 0 && len(roles) > 0 {
		_ = s.roleCache.RefreshPermsCacheByCodes(roles)
		perms, err = s.rolePermsFromCache(ctx, roles)
		if err != nil {
			return nil, errs.SystemError("查询用户权限失败")
		}
	}

	return &model.CurrentUserVO{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
		Perms:    perms,
	}, nil
}

// rolePermsFromCache 从Redis缓存获取角色权限（未命中的角色降级查库）
func (s *Service) rolePermsFromCache(ctx context.Context, roleCodes []string) ([]string, error) {
	if len(roleCodes) == 0 {
		return []string{}, nil
	}

	perms := make([]string, 0)
	missingRoles := make([]string, 0) // 记录缓存中不存在的角色

	for _, roleCode := range roleCodes {
		result, err := redis.Client.HGet(ctx, constant.RedisKeyRolePerms, roleCode).Result()
		if err != nil {
			missingRoles = append(missingRoles, roleCode)
			continue
		}

		if result != "" {
			var rolePerms []string
			if err := json.Unmarshal([]byte(result), &rolePerms); err == nil {
				perms = append(perms, rolePerms...)
			}
		}
	}

	// 降级：缓存未命中时从数据库查询
	if len(missingRoles) > 0 {
		dbPerms, err := s.rolePermsFromDB(ctx, missingRoles)
		if err != nil {
			slog.Error("降级查询数据库失败", "roles", missingRoles, "error", err)
		} else {
			perms = append(perms, dbPerms...)
		}
	}

	return uniqueStrings(perms), nil
}

// rolePermsFromDB 从数据库查询角色权限（降级方案）
func (s *Service) rolePermsFromDB(ctx context.Context, roleCodes []string) ([]string, error) {
	if len(roleCodes) == 0 {
		return []string{}, nil
	}

	rolePermsList, err := s.roleRepo.PermsByCodes(ctx, roleCodes)
	if err != nil {
		return nil, err
	}

	perms := make([]string, 0)
	for _, rolePerms := range rolePermsList {
		perms = append(perms, rolePerms.Perms...)
	}

	return uniqueStrings(perms), nil
}

// Profile 获取个人中心用户信息
func (s *Service) Profile(ctx context.Context, userId int64) (*model.UserProfileVO, error) {
	profile, err := s.repo.Profile(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户信息失败")
	}
	return profile, nil
}

// UpdateProfile 个人中心修改信息（至少修改一项）
func (s *Service) UpdateProfile(ctx context.Context, userId int64, form *model.UserProfileForm) error {
	if form.Nickname == "" && form.Avatar == "" && form.Gender == nil {
		return errs.BadRequest("请至少修改一项")
	}
	if err := s.repo.UpdateProfile(ctx, userId, form); err != nil {
		return errs.SystemError("更新用户信息失败")
	}
	return nil
}

// ResetPassword 重置指定用户密码
func (s *Service) ResetPassword(ctx context.Context, userId int64, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errs.SystemError("密码加密失败")
	}

	if err := s.repo.UpdatePassword(ctx, userId, string(hashedPassword)); err != nil {
		return errs.SystemError("重置密码失败")
	}
	return nil
}

// ChangePassword 当前用户修改密码（校验旧密码、新旧不同）
func (s *Service) ChangePassword(ctx context.Context, userId int64, form *model.PasswordForm) error {
	user, err := s.repo.Get(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return errs.SystemError("查询用户失败")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)); err != nil {
		return errs.BadRequest("旧密码错误")
	}

	if form.NewPassword != form.ConfirmPassword {
		return errs.BadRequest("新密码和确认密码不一致")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.NewPassword)); err == nil {
		return errs.BadRequest("新密码不能与原密码相同")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.SystemError("密码加密失败")
	}

	if err := s.repo.UpdatePassword(ctx, userId, string(hashedPassword)); err != nil {
		return errs.SystemError("修改密码失败")
	}
	return nil
}

// SendMobileCode 发送短信验证码（含发送间隔控制）
func (s *Service) SendMobileCode(ctx context.Context, mobile string) error {
	intervalKey := utils.GetMobileIntervalKey(mobile)
	if err := utils.CheckSendInterval(ctx, intervalKey); err != nil {
		return err
	}

	code := utils.GenerateVerificationCode()

	codeKey := utils.GetMobileCodeKey(mobile)
	if err := utils.StoreVerificationCode(ctx, codeKey, code); err != nil {
		return err
	}

	// TODO: 接入短信服务商并发送验证码
	// smsService.SendSMS(mobile, code)

	slog.Info("短信验证码已发送", "mobile", mobile, "code", code, "expiration_minutes", utils.CodeExpiration)

	return nil
}

// BindOrChangeMobile 绑定或更换手机号（校验密码与验证码，手机号唯一）
func (s *Service) BindOrChangeMobile(ctx context.Context, userId int64, form *model.MobileBindingForm) error {
	if _, err := s.verifyUserAndPassword(ctx, userId, form.Password); err != nil {
		return err
	}

	codeKey := utils.GetMobileCodeKey(form.Mobile)
	if err := utils.VerifyCode(ctx, codeKey, form.Code); err != nil {
		return err
	}

	existingUser, err := s.repo.GetByMobile(ctx, form.Mobile)
	if err == nil && existingUser != nil && existingUser.ID != types.BigInt(userId) {
		return errs.Business("手机号已被其他账号绑定")
	}

	if err := s.repo.UpdateMobile(ctx, userId, form.Mobile); err != nil {
		return errs.SystemError("更新手机号失败")
	}

	return nil
}

// SendEmailCode 发送邮箱验证码（含发送间隔控制）
func (s *Service) SendEmailCode(ctx context.Context, email string) error {
	intervalKey := utils.GetEmailIntervalKey(email)
	if err := utils.CheckSendInterval(ctx, intervalKey); err != nil {
		return err
	}

	code := utils.GenerateVerificationCode()

	codeKey := utils.GetEmailCodeKey(email)
	if err := utils.StoreVerificationCode(ctx, codeKey, code); err != nil {
		return err
	}

	// TODO: 接入 SMTP 或第三方邮件服务
	// emailService.SendEmail(email, "验证码", fmt.Sprintf("您的验证码是：%s", code))

	slog.Info("邮箱验证码已发送", "email", email, "code", code, "expiration_minutes", utils.CodeExpiration)

	return nil
}

// BindOrChangeEmail 绑定或更换邮箱（校验密码与验证码，邮箱唯一）
func (s *Service) BindOrChangeEmail(ctx context.Context, userId int64, form *model.EmailBindingForm) error {
	if _, err := s.verifyUserAndPassword(ctx, userId, form.Password); err != nil {
		return err
	}

	codeKey := utils.GetEmailCodeKey(form.Email)
	if err := utils.VerifyCode(ctx, codeKey, form.Code); err != nil {
		return err
	}

	existingUser, err := s.repo.GetByEmail(ctx, form.Email)
	if err == nil && existingUser != nil && existingUser.ID != types.BigInt(userId) {
		return errs.Business("邮箱已被其他账号绑定")
	}

	if err := s.repo.UpdateEmail(ctx, userId, form.Email); err != nil {
		return errs.SystemError("更新邮箱失败")
	}

	return nil
}

// UnbindMobile 解绑手机号（校验密码，未绑定则拒绝）
func (s *Service) UnbindMobile(ctx context.Context, userId int64, form *model.PasswordVerifyForm) error {
	user, err := s.verifyUserAndPassword(ctx, userId, form.Password)
	if err != nil {
		return err
	}
	if user.Mobile == "" {
		return errs.BadRequest("当前账号未绑定手机号")
	}
	if err := s.repo.UnbindMobile(ctx, userId); err != nil {
		return errs.SystemError("解绑手机号失败")
	}
	return nil
}

// UnbindEmail 解绑邮箱（校验密码，未绑定则拒绝）
func (s *Service) UnbindEmail(ctx context.Context, userId int64, form *model.PasswordVerifyForm) error {
	user, err := s.verifyUserAndPassword(ctx, userId, form.Password)
	if err != nil {
		return err
	}
	if user.Email == "" {
		return errs.BadRequest("当前账号未绑定邮箱")
	}
	if err := s.repo.UnbindEmail(ctx, userId); err != nil {
		return errs.SystemError("解绑邮箱失败")
	}
	return nil
}

// verifyUserAndPassword 校验用户存在性和密码（绑定/解绑共用）
func (s *Service) verifyUserAndPassword(ctx context.Context, userId int64, password string) (*model.User, error) {
	user, err := s.repo.Get(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, errs.BadRequest("当前密码错误")
	}
	return user, nil
}

// Options 用户下拉选项
func (s *Service) Options(ctx context.Context) ([]baseModel.Option[string], error) {
	users, err := s.repo.Options(ctx)
	if err != nil {
		return nil, errs.SystemError("查询用户选项失败")
	}

	options := make([]baseModel.Option[string], len(users))
	for i, user := range users {
		options[i] = baseModel.Option[string]{
			Value: fmt.Sprintf("%d", user.ID),
			Label: user.Nickname,
		}
	}

	return options, nil
}

// ExportToExcel 导出用户数据到Excel（按当前查询条件，上限 ExportMaxLimit）
func (s *Service) ExportToExcel(ctx context.Context, query *model.UserQuery, currentUser *auth.UserDetails) (*excel.ExcelExporter, error) {
	query.PageNum = 1
	query.PageSize = constant.ExportMaxLimit

	users, _, err := s.repo.Page(ctx, query, currentUser)
	if err != nil {
		return nil, errs.SystemError("查询用户数据失败")
	}

	exporter := excel.NewExcelExporter("用户列表")

	headers := []string{
		"用户ID", "用户名", "昵称", "手机号", "性别", "邮箱", "状态", "部门", "角色", "创建时间",
	}
	if err := exporter.SetHeaders(headers); err != nil {
		return nil, errs.SystemError("设置表头失败")
	}

	for _, user := range users {
		gender := map[int]string{0: "未知", 1: "男", 2: "女"}[user.Gender]
		status := map[int]string{0: "禁用", 1: "启用"}[user.Status]

		row := []interface{}{
			user.ID,
			user.Username,
			user.Nickname,
			user.Mobile,
			gender,
			user.Email,
			status,
			user.DeptName,
			user.RoleNames,
			user.CreateTime.String(),
		}
		if err := exporter.AddRow(row); err != nil {
			return nil, errs.SystemError("添加数据行失败")
		}
	}

	return exporter, nil
}

// GenerateTemplate 生成用户导入模板
func (s *Service) GenerateTemplate() (*excel.ExcelExporter, error) {
	exporter := excel.NewExcelExporter("用户导入模板")

	headers := []string{
		"用户名(*)", "昵称(*)", "手机号", "性别(男/女/未知)", "邮箱", "部门ID", "状态(启用/禁用)", "备注",
	}
	if err := exporter.SetHeaders(headers); err != nil {
		return nil, errs.SystemError("设置表头失败")
	}

	examples := [][]interface{}{
		{"zhangsan", "张三", "13800138000", "男", "zhangsan@example.com", "1", "启用", "样例用户1"},
		{"lisi", "李四", "13800138001", "女", "lisi@example.com", "2", "启用", "样例用户2"},
	}

	for _, row := range examples {
		if err := exporter.AddRow(row); err != nil {
			return nil, errs.SystemError("添加样例数据失败")
		}
	}

	return exporter, nil
}

// ImportFromExcel 从Excel导入用户（角色/部门按编码或名称匹配，逐行容错）
func (s *Service) ImportFromExcel(ctx context.Context, file io.Reader) (map[string]interface{}, error) {
	importer, err := excel.NewExcelImporter(file)
	if err != nil {
		return nil, errs.BadRequest("Excel文件格式错误")
	}

	rows, err := importer.GetRows()
	if err != nil {
		return nil, errs.SystemError("读取Excel数据失败")
	}

	if len(rows) < 2 {
		return nil, errs.BadRequest("Excel文件没有数据")
	}

	// 预加载角色和部门数据（支持编码或名称匹配）
	roles, err := s.roleRepo.ListForImport(ctx)
	if err != nil {
		return nil, errs.SystemError("查询角色数据失败")
	}
	roleMap := make(map[string]int64)
	for _, r := range roles {
		if r.Code != "" {
			roleMap[r.Code] = int64(r.ID)
		}
		if r.Name != "" {
			roleMap[r.Name] = int64(r.ID)
		}
	}

	depts, err := s.deptRepo.ListForImport(ctx)
	if err != nil {
		return nil, errs.SystemError("查询部门数据失败")
	}
	deptMap := make(map[string]int64)
	for _, d := range depts {
		if d.Code != "" {
			deptMap[d.Code] = int64(d.ID)
		}
		if d.Name != "" {
			deptMap[d.Name] = int64(d.ID)
		}
	}

	dataRows := rows[1:] // 跳过表头

	successCount := 0
	failCount := 0
	var failDetails []string

	for i, row := range dataRows {
		if len(row) < 2 {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 数据不完整", i+2))
			continue
		}

		username := strings.TrimSpace(row[0])
		nickname := strings.TrimSpace(row[1])
		mobile := ""
		if len(row) > 2 {
			mobile = strings.TrimSpace(row[2])
		}

		genderStr := "未知"
		if len(row) > 3 {
			genderStr = strings.TrimSpace(row[3])
		}
		gender := map[string]int{"男": 1, "女": 2, "未知": 0}[genderStr]

		email := ""
		if len(row) > 4 {
			email = strings.TrimSpace(row[4])
		}

		// 角色列（编码或名称，逗号分隔）
		var roleIds []int64
		if len(row) > 5 && row[5] != "" {
			for _, part := range strings.Split(strings.TrimSpace(row[5]), ",") {
				trimmed := strings.TrimSpace(part)
				if trimmed == "" {
					continue
				}
				if roleId, ok := roleMap[trimmed]; ok {
					roleIds = append(roleIds, roleId)
				}
			}
		}

		// 部门列（编码或名称）
		var deptID int64
		if len(row) > 6 && row[6] != "" {
			if id, ok := deptMap[strings.TrimSpace(row[6])]; ok {
				deptID = id
			}
		}

		status := 1
		if len(row) > 7 && strings.TrimSpace(row[7]) == "禁用" {
			status = 0
		}

		if username == "" || nickname == "" {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 用户名或昵称为空", i+2))
			continue
		}

		exists, _ := s.repo.UsernameExists(ctx, username, 0)
		if exists {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 用户名[%s]已存在", i+2, username))
			continue
		}

		if len(roleIds) == 0 {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 角色不存在或为空", i+2))
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(constant.DefaultPassword), bcrypt.DefaultCost)
		if err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 密码加密失败", i+2))
			continue
		}

		user := &model.User{
			Username: username,
			Nickname: nickname,
			Mobile:   mobile,
			Gender:   gender,
			Email:    email,
			DeptID:   types.BigInt(deptID),
			Status:   status,
			Password: string(hashedPassword),
		}

		if err := s.repo.Create(ctx, user); err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 创建失败 - %v", i+2, err))
			continue
		}

		if err := s.repo.UpdateRoles(ctx, int64(user.ID), roleIds); err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 分配角色失败 - %v", i+2, err))
			continue
		}

		successCount++
	}

	return map[string]interface{}{
		"total":       len(dataRows),
		"success":     successCount,
		"fail":        failCount,
		"failDetails": failDetails,
	}, nil
}

// uniqueStrings 去重字符串切片（忽略空串）
func uniqueStrings(strs []string) []string {
	if len(strs) == 0 {
		return []string{}
	}
	set := make(map[string]bool)
	for _, s := range strs {
		if s != "" {
			set[s] = true
		}
	}
	result := make([]string, 0, len(set))
	for s := range set {
		result = append(result, s)
	}
	return result
}
