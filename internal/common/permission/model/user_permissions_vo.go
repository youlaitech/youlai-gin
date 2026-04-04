package model

import (
	"youlai-gin/internal/common/auth"
	"youlai-gin/pkg/types"
)

// UserPermissionsVO 用户权限信息
type UserPermissionsVO struct {
	UserID     types.BigInt         `json:"userId"`
	Roles      []string             `json:"roles"`
	Perms      []string             `json:"perms"`
	DataScopes []auth.RoleDataScope `json:"dataScopes"`
	DeptID     types.BigInt         `json:"deptId"`
}
