package model

import "youlai-gin/pkg/types"

// UserPermissionsVO 用户权限信息（供权限校验、数据权限过滤使用）
type UserPermissionsVO struct {
	UserID    types.BigInt   `json:"userId"`
	Roles     []string       `json:"roles"`
	Perms     []string       `json:"perms"`
	DataScope int            `json:"dataScope"`
	DeptID    types.BigInt   `json:"deptId"`
	DeptIds   []types.BigInt `json:"deptIds"`
}
