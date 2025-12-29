package model

import "youlai-gin/pkg/types"

// UserPageVO 用户分页视图对象
type UserPageVO struct {
	ID         types.BigInt     `json:"id"`
	Username   string           `json:"username"`
	Nickname   string           `json:"nickname"`
	Mobile     string           `json:"mobile"`
	Gender     int              `json:"gender"`
	Avatar     string           `json:"avatar"`
	Email      string           `json:"email"`
	Status     int              `json:"status"`
	DeptName   string           `json:"deptName"`
	RoleNames  string           `json:"roleNames"`
	CreateTime types.LocalTime  `json:"createTime"`
}

// UserProfileVO 个人中心用户信息
type UserProfileVO struct {
	ID       types.BigInt `json:"id"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Gender   int          `json:"gender"`
	Mobile   string       `json:"mobile"`
	Email    string       `json:"email"`
	DeptName string       `json:"deptName"`
	RoleNames string      `json:"roleNames"`
}

// CurrentUserDTO 当前登录用户信息
type CurrentUserDTO struct {
	UserID   types.BigInt `json:"userId"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Roles    []string     `json:"roles"`
	Perms    []string     `json:"perms"`
}
