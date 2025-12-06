package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youlai-gin/internal/system/user/model"
	"youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
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

// GetCurrentUserInfo 获取当前登录用户信息
func GetCurrentUserInfo(userId int64) (*model.CurrentUserDTO, error) {
	user, err := repository.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// TODO: 获取用户角色和权限
	return &model.CurrentUserDTO{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    []string{},
		Perms:    []string{},
	}, nil
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
