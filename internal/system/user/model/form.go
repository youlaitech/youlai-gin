package model

import "youlai-gin/pkg/types"

// UserForm 用户新增/更新表单
type UserForm struct {
	ID       types.BigInt   `json:"id"`
	Username string         `json:"username" binding:"required"`
	Nickname string         `json:"nickname" binding:"required"`
	Mobile   string         `json:"mobile"`
	Gender   int            `json:"gender"`
	Avatar   string         `json:"avatar"`
	Email    string         `json:"email"`
	Status   int            `json:"status"`
	DeptID   types.BigInt   `json:"deptId"`
	RoleIDs  []types.BigInt `json:"roleIds" binding:"required"`
	Openid   string         `json:"openId"`
}

// UserProfileForm 个人中心用户信息更新表单
type UserProfileForm struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   *int   `json:"gender"`
}

// PasswordForm 修改密码表单
type PasswordForm struct {
	OldPassword     string `json:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// MobileBindingForm 绑定或更换手机号表单
type MobileBindingForm struct {
	Mobile   string `json:"mobile" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// EmailBindingForm 绑定或更换邮箱表单
type EmailBindingForm struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// PasswordVerifyForm 密码校验表单（解绑手机号/邮箱使用）
type PasswordVerifyForm struct {
	Password string `json:"password" binding:"required"`
}
