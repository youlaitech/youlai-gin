package model

import "youlai-gin/pkg/types"

// UserPageVO 用户分页视图
type UserPageVO struct {
	ID         types.BigInt    `json:"id"` // 主键
	Username   string          `json:"username"` // 用户名
	Nickname   string          `json:"nickname"` // 昵称
	Mobile     string          `json:"mobile"` // 手机号
	Gender     int             `json:"gender"` // 性别
	Avatar     string          `json:"avatar"` // 头像
	Email      string          `json:"email"` // 邮箱
	Status     int             `json:"status"` // 状态(1启用0禁用)
	DeptName   string          `json:"deptName"`
	RoleNames  string          `json:"roleNames"`
	CreateTime types.LocalTime `json:"createTime"` // 创建时间
}

// UserProfileVO 个人中心用户信息视图
type UserProfileVO struct {
	ID        types.BigInt `json:"id"` // 主键
	Username  string       `json:"username"` // 用户名
	Nickname  string       `json:"nickname"` // 昵称
	Avatar    string       `json:"avatar"` // 头像
	Gender    int          `json:"gender"` // 性别
	Mobile    string       `json:"mobile"` // 手机号
	Email     string       `json:"email"` // 邮箱
	DeptName  string       `json:"deptName"`
	RoleNames string       `json:"roleNames"`
}

// CurrentUserVO 当前登录用户信息视图
type CurrentUserVO struct {
	UserID   types.BigInt `json:"userId"`
	Username string       `json:"username"` // 用户名
	Nickname string       `json:"nickname"` // 昵称
	Avatar   string       `json:"avatar"` // 头像
	Roles    []string     `json:"roles"` // 角色列表
	Perms    []string     `json:"perms"` // 权限列表
}

// UserFormVO 用户表单视图
type UserFormVO struct {
	ID       types.BigInt   `json:"id"` // 主键
	Username string         `json:"username"` // 用户名
	Nickname string         `json:"nickname"` // 昵称
	Mobile   string         `json:"mobile"` // 手机号
	Gender   int            `json:"gender"` // 性别
	Avatar   string         `json:"avatar"` // 头像
	Email    string         `json:"email"` // 邮箱
	Status   int            `json:"status"` // 状态(1启用0禁用)
	DeptID   types.BigInt   `json:"deptId"` // 部门ID
	RoleIDs  []types.BigInt `json:"roleIds"`
	Openid   string         `json:"openId"`
}
