package model

import "youlai-gin/pkg/types"

// UserPageVO 用户分页视图
type UserPageVO struct {
	ID         types.BigInt    `json:"id"`
	Username   string          `json:"username"`
	Nickname   string          `json:"nickname"`
	Mobile     string          `json:"mobile"`
	Gender     int             `json:"gender"`
	Avatar     string          `json:"avatar"`
	Email      string          `json:"email"`
	Status     int             `json:"status"`
	DeptName   string          `json:"deptName"`
	RoleNames  string          `json:"roleNames"`
	CreateTime types.LocalTime `json:"createTime"`
}

// UserProfileVO 个人中心用户信息视图
type UserProfileVO struct {
	ID        types.BigInt `json:"id"`
	Username  string       `json:"username"`
	Nickname  string       `json:"nickname"`
	Avatar    string       `json:"avatar"`
	Gender    int          `json:"gender"`
	Mobile    string       `json:"mobile"`
	Email     string       `json:"email"`
	DeptName  string       `json:"deptName"`
	RoleNames string       `json:"roleNames"`
}

// CurrentUserVO 当前登录用户信息视图
type CurrentUserVO struct {
	UserID   types.BigInt `json:"userId"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Roles    []string     `json:"roles"`
	Perms    []string     `json:"perms"`
}

// UserFormVO 用户表单视图
type UserFormVO struct {
	ID       types.BigInt   `json:"id"`
	Username string         `json:"username"`
	Nickname string         `json:"nickname"`
	Mobile   string         `json:"mobile"`
	Gender   int            `json:"gender"`
	Avatar   string         `json:"avatar"`
	Email    string         `json:"email"`
	Status   int            `json:"status"`
	DeptID   types.BigInt   `json:"deptId"`
	RoleIDs  []types.BigInt `json:"roleIds"`
	Openid   string         `json:"openId"`
}
