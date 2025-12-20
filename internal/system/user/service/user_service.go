package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youlai-gin/internal/system/user/model"
	"youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/excel"
	"youlai-gin/pkg/redis"
)

// GetUserPage 用户分页列表
func GetUserPage(query *model.UserPageQuery) (*common.PageResult, error) {
	users, total, err := repository.GetUserPage(query)
	if err != nil {
		return nil, errs.SystemError("查询用户列表失败")
	}

	return &common.PageResult{
		List:  users,
		Total: total,
	}, nil
}

// SaveUser 保存用户（新增或更新）
func SaveUser(form *model.UserForm) error {
	// 检查用户名是否存在
	exists, err := repository.CheckUsernameExists(form.Username, form.ID)
	if err != nil {
		return errs.SystemError("检查用户名失败")
	}
	if exists {
		return errs.BadRequest("用户名已存在")
	}

	user := &model.User{
		ID:       form.ID,
		Username: form.Username,
		Nickname: form.Nickname,
		Mobile:   form.Mobile,
		Gender:   form.Gender,
		Avatar:   form.Avatar,
		Email:    form.Email,
		Status:   form.Status,
		DeptID:   form.DeptID,
		Openid:   form.Openid,
	}

	// 新增用户需要设置默认密码
	if form.ID == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if err != nil {
			return errs.SystemError("密码加密失败")
		}
		user.Password = string(hashedPassword)

		if err := repository.CreateUser(user); err != nil {
			return errs.SystemError("创建用户失败")
		}
	} else {
		if err := repository.UpdateUser(user); err != nil {
			return errs.SystemError("更新用户失败")
		}
	}

	// 保存用户角色关联
	if err := repository.SaveUserRoles(user.ID, form.RoleIDs); err != nil {
		return errs.SystemError("保存用户角色失败")
	}

	return nil
}

// GetUserForm 获取用户表单数据
func GetUserForm(id int64) (*model.UserForm, error) {
	user, err := repository.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 获取用户角色ID列表
	roleIds, err := repository.GetUserRoleIDs(id)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	return &model.UserForm{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Status:   user.Status,
		DeptID:   user.DeptID,
		RoleIDs:  roleIds,
		Openid:   user.Openid,
	}, nil
}

// DeleteUsers 批量删除用户
func DeleteUsers(idsStr string) error {
	if idsStr == "" {
		return errs.BadRequest("用户ID不能为空")
	}

	idStrs := strings.Split(idsStr, ",")
	ids := make([]int64, 0, len(idStrs))
	for _, idStr := range idStrs {
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return errs.BadRequest("无效的用户ID")
	}

	if err := repository.DeleteUsersByIDs(ids); err != nil {
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

// GetCurrentUserInfo 获取当前登录用户信息（需要传入token中的userDetails）
func GetCurrentUserInfoWithRoles(userId int64, roles []string) (*model.CurrentUserDTO, error) {
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

	return &model.CurrentUserDTO{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
		Perms:    perms,
	}, nil
}

// GetCurrentUserInfo 获取当前登录用户信息（从数据库查询角色）
func GetCurrentUserInfo(userId int64) (*model.CurrentUserDTO, error) {
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

	return &model.CurrentUserDTO{
		UserID:   user.ID,
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
func GetUserProfile(userId int64) (*model.UserProfileVO, error) {
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
func UpdateUserProfile(userId int64, form *model.UserProfileForm) error {
	if err := repository.UpdateUserProfile(userId, form); err != nil {
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
func ChangeUserPassword(userId int64, form *model.PasswordUpdateForm) error {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		return errs.SystemError("查询用户失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)); err != nil {
		return errs.BadRequest("旧密码错误")
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
	// TODO: 实现短信验证码发送逻辑
	return nil
}

// BindOrChangeMobile 绑定或更换手机号
func BindOrChangeMobile(userId int64, form *model.MobileUpdateForm) error {
	// TODO: 验证短信验证码

	if err := repository.UpdateUserMobile(userId, form.Mobile); err != nil {
		return errs.SystemError("更新手机号失败")
	}
	return nil
}

// SendEmailCode 发送邮箱验证码
func SendEmailCode(email string) error {
	// TODO: 实现邮箱验证码发送逻辑
	return nil
}

// BindOrChangeEmail 绑定或更换邮箱
func BindOrChangeEmail(userId int64, form *model.EmailUpdateForm) error {
	// TODO: 验证邮箱验证码

	if err := repository.UpdateUserEmail(userId, form.Email); err != nil {
		return errs.SystemError("更新邮箱失败")
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
func ExportUsersToExcel(query *model.UserPageQuery) (*excel.ExcelExporter, error) {
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
		user := &model.User{
			Username: username,
			Nickname: nickname,
			Mobile:   mobile,
			Gender:   gender,
			Email:    email,
			DeptID:   deptID,
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
