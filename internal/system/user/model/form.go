package model

// UserForm 用户表单
type UserForm struct {
	ID       int64   `json:"id"`
	Username string  `json:"username" binding:"required"`
	Nickname string  `json:"nickname" binding:"required"`
	Mobile   string  `json:"mobile"`
	Gender   int     `json:"gender"`
	Avatar   string  `json:"avatar"`
	Email    string  `json:"email"`
	Status   int     `json:"status"`
	DeptID   int64   `json:"deptId"`
	RoleIDs  []int64 `json:"roleIds" binding:"required"`
	Openid   string  `json:"openId"`
}

// UserProfileForm 个人中心用户信息表单
type UserProfileForm struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int    `json:"gender"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

// PasswordUpdateForm 修改密码表单
type PasswordUpdateForm struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// MobileUpdateForm 绑定或更换手机号表单
type MobileUpdateForm struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

// EmailUpdateForm 绑定或更换邮箱表单
type EmailUpdateForm struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}
