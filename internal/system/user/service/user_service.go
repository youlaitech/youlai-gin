package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youlai-gin/internal/system/user/api"
	"youlai-gin/internal/system/user/domain"
	"youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/excel"
	"youlai-gin/pkg/redis"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/utils"
)

// GetUserPage 用户分页列表
func GetUserPage(query *api.UserQueryReq) (*common.PagedData, error) {
	users, total, err := repository.GetUserPage(query)
	if err != nil {
		return nil, errs.SystemError("查询用户列表失败")
	}

	pageMeta := common.NewPageMeta(query.PageNum, query.PageSize, total)
	return &common.PagedData{Data: users, Page: pageMeta}, nil
}

// SaveUser 保存用户（新增或更新）
func SaveUser(form *api.UserSaveReq) error {
	// 检查用户名是否已存在
	exists, err := repository.CheckUsernameExists(form.Username, int64(form.ID))
	if err != nil {
		return errs.SystemError("检查用户名失败")
	}
	if exists {
		return errs.New(constant.CodeUserRegistrationError, "用户名已存在", http.StatusBadRequest)
	}

	// 转换为实体
	user := &domain.User{
		Username: form.Username,
		Nickname: form.Nickname,
		Mobile:   form.Mobile,
		Gender:   form.Gender,
		Email:    form.Email,
		DeptID:   form.DeptID,
		Status:   form.Status,
		Avatar:   form.Avatar,
	}

	if form.ID > 0 {
		// 更新用户
		user.ID = types.BigInt(int64(form.ID))
		if err := repository.UpdateUser(user); err != nil {
			return errs.SystemError("更新用户失败")
		}

		// 更新用户角色
		roleIDs := make([]int64, len(form.RoleIDs))
		for i, roleID := range form.RoleIDs {
			roleIDs[i] = int64(roleID)
		}
		if err := repository.SaveUserRoles(int64(form.ID), roleIDs); err != nil {
			return errs.SystemError("更新用户角色失败")
		}
	} else {
		// 创建用户 - 设置默认密码
		defaultPassword := "123456"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return errs.SystemError("密码加密失败")
		}
		user.Password = string(hashedPassword)

		if err := repository.CreateUser(user); err != nil {
			return errs.SystemError("创建用户失败")
		}

		// 分配角色
		if len(form.RoleIDs) > 0 {
			roleIDs := make([]int64, len(form.RoleIDs))
			for i, roleID := range form.RoleIDs {
				roleIDs[i] = int64(roleID)
			}
			if err := repository.SaveUserRoles(int64(user.ID), roleIDs); err != nil {
				return errs.SystemError("分配用户角色失败")
			}
		}
	}

	return nil
}

// GetUserForm 获取用户表单数据
func GetUserForm(userId int64) (*api.UserFormResp, error) {
	if userId == 0 {
		// 新增用户，返回空表单
		return &api.UserFormResp{}, nil
	}

	// 查询用户信息
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 查询用户角色ID列表
	roleIDs, err := repository.GetUserRoleIDs(userId)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	// 转换角色ID类型
	bigIntRoleIDs := make([]types.BigInt, len(roleIDs))
	for i, roleID := range roleIDs {
		bigIntRoleIDs[i] = types.BigInt(roleID)
	}

	return &api.UserFormResp{
		ID:       types.BigInt(user.ID),
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Email:    user.Email,
		Avatar:   user.Avatar,
		DeptID:   user.DeptID,
		Status:   user.Status,
		RoleIDs:  bigIntRoleIDs,
	}, nil
}

// DeleteUsers 删除用户
func DeleteUsers(ids string) error {
	if ids == "" {
		return errs.BadRequest("请选择要删除的用户")
	}

	// 解析ID列表
	idList := strings.Split(ids, ",")
	userIDs := make([]int64, 0, len(idList))
	for _, idStr := range idList {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		return errs.BadRequest("无效的用户ID")
	}

	// 删除用户
	if err := repository.DeleteUsersByIDs(userIDs); err != nil {
		return errs.SystemError("删除用户失败")
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(userId int64, status int) error {
	if err := repository.UpdateUserStatus(userId, status); err != nil {
		return errs.SystemError("更新用户状态失败")
	}
	return nil
}

// GetCurrentUserInfoWithRoles 获取当前登录用户信息（需要传入token中的userDetails）
func GetCurrentUserInfoWithRoles(userId int64, roles []string) (*api.CurrentUserResp, error) {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 获取用户权限列表（从Redis缓存）
	perms := []string{}
	if len(roles) > 0 {
		perms, err = getRolePermsFromCache(roles)
		if err != nil {
			return nil, errs.SystemError("查询用户权限失败")
		}
	}

	return &api.CurrentUserResp{
		UserID:   types.BigInt(user.ID),
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
		Perms:    perms,
	}, nil
}

// GetCurrentUserInfo 获取当前登录用户信息（从数据库查询角色）
func GetCurrentUserInfo(userId int64) (*api.CurrentUserResp, error) {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 从数据库获取用户角色编码列表
	roles, err := repository.GetUserRoles(userId)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	// 获取用户权限列表
	perms := []string{}
	if len(roles) > 0 {
		perms, err = getRolePermsFromCache(roles)
		if err != nil {
			return nil, errs.SystemError("查询用户权限失败")
		}
	}

	return &api.CurrentUserResp{
		UserID:   types.BigInt(user.ID),
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
		Perms:    perms,
	}, nil
}

// getRolePermsFromCache 从Redis缓存中获取角色权限列表（带降级策略）
func getRolePermsFromCache(roleCodes []string) ([]string, error) {
	if len(roleCodes) == 0 {
		return []string{}, nil
	}

	ctx := context.Background()
	permsSet := make(map[string]bool)
	missingRoles := make([]string, 0) // 记录缓存中不存在的角色

	// 从Redis中获取每个角色的权限
	for _, roleCode := range roleCodes {
		// Redis key: system:role:perms
		result, err := redis.Client.HGet(ctx, "system:role:perms", roleCode).Result()
		if err != nil {
			// 记录缓存未命中的角色，稍后降级查询数据库
			missingRoles = append(missingRoles, roleCode)
			continue
		}

		if result != "" {
			// 尝试解析JSON数组格式
			var rolePerms []string
			if err := json.Unmarshal([]byte(result), &rolePerms); err == nil {
				for _, perm := range rolePerms {
					if perm != "" {
						permsSet[perm] = true
					}
				}
			}
		}
	}

	// 降级策略：如果有角色在缓存中不存在，从数据库查询
	if len(missingRoles) > 0 {
		dbPerms, err := getRolePermsFromDB(missingRoles)
		if err != nil {
			// 数据库查询失败，只返回已从缓存获取的权限
			// 不返回错误，保证服务可用性
			fmt.Printf("⚠️  降级查询数据库失败，角色: %v, 错误: %v\n", missingRoles, err)
		} else {
			// 将数据库查询结果合并到权限集合
			for _, perm := range dbPerms {
				if perm != "" {
					permsSet[perm] = true
				}
			}
		}
	}

	// 将set转为slice
	perms := make([]string, 0, len(permsSet))
	for perm := range permsSet {
		perms = append(perms, perm)
	}

	return perms, nil
}

// getRolePermsFromDB 从数据库查询角色权限（降级方案）
func getRolePermsFromDB(roleCodes []string) ([]string, error) {
	if len(roleCodes) == 0 {
		return []string{}, nil
	}

	// 导入role repository（避免循环依赖，使用数据库直接查询）
	rolePermsList, err := repository.GetRolePermsByCodes(roleCodes)
	if err != nil {
		return nil, err
	}

	// 收集所有权限
	permsSet := make(map[string]bool)
	for _, rolePerms := range rolePermsList {
		for _, perm := range rolePerms.Perms {
			if perm != "" {
				permsSet[perm] = true
			}
		}
	}

	// 转为slice
	perms := make([]string, 0, len(permsSet))
	for perm := range permsSet {
		perms = append(perms, perm)
	}

	return perms, nil
}

// GetUserProfile 获取用户个人信息
func GetUserProfile(userId int64) (*api.UserProfileResp, error) {
	profile, err := repository.GetUserProfile(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户信息失败")
	}
	return profile, nil
}

// UpdateUserProfile 更新用户个人信息
func UpdateUserProfile(userId int64, req *api.UserProfileUpdateReq) error {
	if req.Nickname == "" && req.Avatar == "" && req.Gender == nil {
		return errs.BadRequest("请至少修改一项")
	}
	if err := repository.UpdateUserProfile(userId, req); err != nil {
		return errs.SystemError("更新用户信息失败")
	}
	return nil
}

// ResetUserPassword 重置用户密码
func ResetUserPassword(userId int64, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errs.SystemError("密码加密失败")
	}

	if err := repository.UpdateUserPassword(userId, string(hashedPassword)); err != nil {
		return errs.SystemError("重置密码失败")
	}
	return nil
}

// ChangeUserPassword 当前用户修改密码
func ChangeUserPassword(userId int64, form *api.PasswordUpdateReq) error {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		return errs.SystemError("查询用户失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)); err != nil {
		return errs.BadRequest("旧密码错误")
	}

	if form.NewPassword != form.ConfirmPassword {
		return errs.BadRequest("新密码和确认密码不一致")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.NewPassword)); err == nil {
		return errs.BadRequest("新密码不能与原密码相同")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.SystemError("密码加密失败")
	}

	if err := repository.UpdateUserPassword(userId, string(hashedPassword)); err != nil {
		return errs.SystemError("修改密码失败")
	}
	return nil
}

// SendMobileCode 发送短信验证码
func SendMobileCode(mobile string) error {
	ctx := context.Background()

	// 1. 检查发送间隔
	intervalKey := utils.GetMobileIntervalKey(mobile)
	if err := utils.CheckSendInterval(ctx, intervalKey); err != nil {
		return err
	}

	// 2. 生成验证码
	code := utils.GenerateVerificationCode()

	// 3. 存储验证码到 Redis
	codeKey := utils.GetMobileCodeKey(mobile)
	if err := utils.StoreVerificationCode(ctx, codeKey, code); err != nil {
		return err
	}

	// 4. 发送短信（实际生产环境对接短信服务商）
	// TODO: 对接阿里云、腾讯云等短信服务
	// 示例：smsService.SendSMS(mobile, code)

	// 开发环境：打印验证码到日志
	fmt.Printf("📱 短信验证码已发送到 %s: %s (有效期 %d 分钟)\n", mobile, code, utils.CodeExpiration)

	return nil
}

// BindOrChangeMobile 绑定或更换手机号
func BindOrChangeMobile(userId int64, form *api.MobileUpdateReq) error {
	ctx := context.Background()

	// 0. 校验当前密码
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return errs.SystemError("查询用户失败")
	}
	if err := utils.VerifyPassword(user.Password, form.Password); err != nil {
		return errs.BadRequest("当前密码错误")
	}

	// 1. 验证短信验证码
	codeKey := utils.GetMobileCodeKey(form.Mobile)
	if err := utils.VerifyCode(ctx, codeKey, form.Code); err != nil {
		return err
	}

	// 2. 检查手机号是否已被其他用户使用
	existingUser, err := repository.GetUserByMobile(form.Mobile)
	if err == nil && existingUser != nil && existingUser.ID != types.BigInt(userId) {
		return errs.BadRequest("手机号已被其他账号绑定")
	}

	// 3. 更新手机号
	if err := repository.UpdateUserMobile(userId, form.Mobile); err != nil {
		return errs.SystemError("更新手机号失败")
	}

	return nil
}

// SendEmailCode 发送邮箱验证码
func SendEmailCode(email string) error {
	ctx := context.Background()

	// 1. 检查发送间隔
	intervalKey := utils.GetEmailIntervalKey(email)
	if err := utils.CheckSendInterval(ctx, intervalKey); err != nil {
		return err
	}

	// 2. 生成验证码
	code := utils.GenerateVerificationCode()

	// 3. 存储验证码到 Redis
	codeKey := utils.GetEmailCodeKey(email)
	if err := utils.StoreVerificationCode(ctx, codeKey, code); err != nil {
		return err
	}

	// 4. 发送邮件（实际生产环境对接邮件服务）
	// TODO: 对接 SMTP 服务或第三方邮件服务
	// 示例：emailService.SendEmail(email, "验证码", fmt.Sprintf("您的验证码是：%s", code))

	// 开发环境：打印验证码到日志
	fmt.Printf("📧 邮箱验证码已发送到 %s: %s (有效期 %d 分钟)\n", email, code, utils.CodeExpiration)

	return nil
}

// BindOrChangeEmail 绑定或更换邮箱
func BindOrChangeEmail(userId int64, form *api.EmailUpdateReq) error {
	ctx := context.Background()

	// 0. 校验当前密码
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return errs.SystemError("查询用户失败")
	}
	if err := utils.VerifyPassword(user.Password, form.Password); err != nil {
		return errs.BadRequest("当前密码错误")
	}

	// 1. 验证邮箱验证码
	codeKey := utils.GetEmailCodeKey(form.Email)
	if err := utils.VerifyCode(ctx, codeKey, form.Code); err != nil {
		return err
	}

	// 2. 检查邮箱是否已被其他用户使用
	existingUser, err := repository.GetUserByEmail(form.Email)
	if err == nil && existingUser != nil && existingUser.ID != types.BigInt(userId) {
		return errs.BadRequest("邮箱已被其他账号绑定")
	}

	// 3. 更新邮箱
	if err := repository.UpdateUserEmail(userId, form.Email); err != nil {
		return errs.SystemError("更新邮箱失败")
	}

	return nil
}

// UnbindMobile 解绑手机号
func UnbindMobile(userId int64, form *api.PasswordVerifyReq) error {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return errs.SystemError("查询用户失败")
	}
	if user.Mobile == "" {
		return errs.BadRequest("当前账号未绑定手机号")
	}
	if err := utils.VerifyPassword(user.Password, form.Password); err != nil {
		return errs.BadRequest("当前密码错误")
	}
	if err := repository.UnbindUserMobile(userId); err != nil {
		return errs.SystemError("解绑手机号失败")
	}
	return nil
}

// UnbindEmail 解绑邮箱
func UnbindEmail(userId int64, form *api.PasswordVerifyReq) error {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return errs.SystemError("查询用户失败")
	}
	if user.Email == "" {
		return errs.BadRequest("当前账号未绑定邮箱")
	}
	if err := utils.VerifyPassword(user.Password, form.Password); err != nil {
		return errs.BadRequest("当前密码错误")
	}
	if err := repository.UnbindUserEmail(userId); err != nil {
		return errs.SystemError("解绑邮箱失败")
	}
	return nil
}

// GetUserOptions 获取用户下拉选项
func GetUserOptions() ([]common.Option[string], error) {
	users, err := repository.GetUserOptions()
	if err != nil {
		return nil, errs.SystemError("查询用户选项失败")
	}

	options := make([]common.Option[string], len(users))
	for i, user := range users {
		options[i] = common.Option[string]{
			Value: fmt.Sprintf("%d", user.ID),
			Label: user.Nickname,
		}
	}

	return options, nil
}

// ExportUsersToExcel 导出用户数据到Excel
func ExportUsersToExcel(query *api.UserQueryReq) (*excel.ExcelExporter, error) {
	// 查询所有符合条件的用户（不分页）
	query.PageNum = 1
	query.PageSize = 10000 // 设置一个较大的值

	users, _, err := repository.GetUserPage(query)
	if err != nil {
		return nil, errs.SystemError("查询用户数据失败")
	}

	// 创建Excel导出器
	exporter := excel.NewExcelExporter("用户列表")

	// 设置表头
	headers := []string{
		"用户ID", "用户名", "昵称", "手机号", "性别", "邮箱", "状态", "部门", "角色", "创建时间",
	}
	if err := exporter.SetHeaders(headers); err != nil {
		return nil, errs.SystemError("设置表头失败")
	}

	// 添加数据行
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

// GenerateUserTemplate 生成用户导入模板
func GenerateUserTemplate() (*excel.ExcelExporter, error) {
	exporter := excel.NewExcelExporter("用户导入模板")

	// 设置表头
	headers := []string{
		"用户名(*)", "昵称(*)", "手机号", "性别(男/女/未知)", "邮箱", "部门ID", "状态(启用/禁用)", "备注",
	}
	if err := exporter.SetHeaders(headers); err != nil {
		return nil, errs.SystemError("设置表头失败")
	}

	// 添加示例数据行
	examples := [][]interface{}{
		{"zhangsan", "张三", "13800138000", "男", "zhangsan@example.com", "1", "启用", "示例用户1"},
		{"lisi", "李四", "13800138001", "女", "lisi@example.com", "2", "启用", "示例用户2"},
	}

	for _, row := range examples {
		if err := exporter.AddRow(row); err != nil {
			return nil, errs.SystemError("添加示例数据失败")
		}
	}

	return exporter, nil
}

// ImportUsersFromExcel 从Excel导入用户数据
func ImportUsersFromExcel(file io.Reader) (map[string]interface{}, error) {
	importer, err := excel.NewExcelImporter(file)
	if err != nil {
		return nil, errs.BadRequest("Excel文件格式错误")
	}
	defer importer.Close()

	rows, err := importer.GetRows()
	if err != nil {
		return nil, errs.SystemError("读取Excel数据失败")
	}

	if len(rows) < 2 {
		return nil, errs.BadRequest("Excel文件没有数据")
	}

	// 跳过表头
	dataRows := rows[1:]

	successCount := 0
	failCount := 0
	var failDetails []string

	for i, row := range dataRows {
		if len(row) < 2 {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 数据不完整", i+2))
			continue
		}

		// 解析行数据
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

		deptID := int64(0)
		if len(row) > 5 && row[5] != "" {
			deptIDVal, _ := strconv.ParseInt(strings.TrimSpace(row[5]), 10, 64)
			deptID = deptIDVal
		}

		status := 1
		if len(row) > 6 {
			statusStr := strings.TrimSpace(row[6])
			if statusStr == "禁用" {
				status = 0
			}
		}

		// 验证必填字段
		if username == "" || nickname == "" {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 用户名或昵称为空", i+2))
			continue
		}

		// 检查用户名是否已存在
		exists, _ := repository.CheckUsernameExists(username, 0)
		if exists {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 用户名[%s]已存在", i+2, username))
			continue
		}

		// 创建用户
		user := &domain.User{
			Username: username,
			Nickname: nickname,
			Mobile:   mobile,
			Gender:   gender,
			Email:    email,
			DeptID:   types.BigInt(deptID),
			Status:   status,
			Password: "$2a$10$xqb1QjFdvVXMHrdLHKHgG.SQWZpfqnLSQEDdE/eUcLfnXW6rMaLTK", // 默认密码: 123456
		}

		if err := repository.CreateUser(user); err != nil {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("第%d行: 创建失败 - %v", i+2, err))
			continue
		}

		successCount++
	}

	result := map[string]interface{}{
		"total":       len(dataRows),
		"success":     successCount,
		"fail":        failCount,
		"failDetails": failDetails,
	}

	return result, nil
}
