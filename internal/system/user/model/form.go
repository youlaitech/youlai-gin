package model

import "youlai-gin/pkg/types"

// UserForm 用户新增/更新表单
// RoleIDs/Openid 不是 sys_user 表字段，用 gorm:"-" 排除，由 Service/Repository 单独处理。
type UserForm struct {
	ID       types.BigInt   `json:"id"` // 主键
	Username string         `json:"username" binding:"required"` // 用户名
	Nickname string         `json:"nickname" binding:"required"` // 昵称
	Mobile   string         `json:"mobile"` // 手机号
	Gender   types.FlexInt  `json:"gender"` // 性别 (兼容字符串和数字)
	Avatar   string         `json:"avatar"` // 头像
	Email    string         `json:"email"` // 邮箱
	Status   types.FlexInt  `json:"status"` // 状态(1启用0禁用) (兼容字符串和数字)
	DeptID   types.BigInt   `json:"deptId"` // 部门ID
	RoleIDs  []types.BigInt `json:"roleIds" binding:"required" gorm:"-"` // 角色ID（关联关系，非表字段）
	Openid   string         `json:"openId" gorm:"-"` // 第三方openid（非表字段，落库到 user_social）
}

// UserProfileForm 个人中心用户信息更新表单
type UserProfileForm struct {
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"` // 头像
	Gender   *types.FlexInt `json:"gender"` // 性别 (兼容字符串和数字)
}

// PasswordForm 修改密码表单
type PasswordForm struct {
	OldPassword     string `json:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// MobileBindingForm 绑定或更换手机号表单
type MobileBindingForm struct {
	Mobile   string `json:"mobile" binding:"required"` // 手机号
	Code     string `json:"code" binding:"required"` // 编码
	Password string `json:"password" binding:"required"` // 密码
}

// EmailBindingForm 绑定或更换邮箱表单
type EmailBindingForm struct {
	Email    string `json:"email" binding:"required"` // 邮箱
	Code     string `json:"code" binding:"required"` // 编码
	Password string `json:"password" binding:"required"` // 密码
}

// PasswordVerifyForm 密码校验表单（解绑手机号/邮箱使用）
type PasswordVerifyForm struct {
	Password string `json:"password" binding:"required"` // 密码
}
