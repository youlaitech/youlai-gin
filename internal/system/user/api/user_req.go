package api

import "youlai-gin/pkg/types"

// UserSaveReq 用户新增/更新请求
type UserSaveReq struct {
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

// UserProfileUpdateReq 个人中心用户信息更新请求
type UserProfileUpdateReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   *int   `json:"gender"`
}

// PasswordUpdateReq 修改密码请求
type PasswordUpdateReq struct {
	OldPassword     string `json:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// MobileUpdateReq 绑定或更换手机号请求
type MobileUpdateReq struct {
	Mobile   string `json:"mobile" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// EmailUpdateReq 绑定或更换邮箱请求
type EmailUpdateReq struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// PasswordVerifyReq 密码校验请求（解绑手机号/邮箱使用）
type PasswordVerifyReq struct {
	Password string `json:"password" binding:"required"`
}
