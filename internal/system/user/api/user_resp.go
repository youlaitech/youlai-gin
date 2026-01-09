package api

import "youlai-gin/pkg/types"

// UserPageResp 用户分页响应
type UserPageResp struct {
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

// UserProfileResp 个人中心用户信息响应
type UserProfileResp struct {
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

// CurrentUserResp 当前登录用户信息响应
type CurrentUserResp struct {
	UserID   types.BigInt `json:"userId"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Roles    []string     `json:"roles"`
	Perms    []string     `json:"perms"`
}

// UserFormResp 用户表单响应
type UserFormResp struct {
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
