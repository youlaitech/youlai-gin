package model

import (
	"youlai-gin/pkg/types"
)

// RoleDataScope 角色数据权限信息，支持多角色数据权限合并（并集策略）
type RoleDataScope struct {
	RoleCode      string  `json:"roleCode"`
	DataScope     int     `json:"dataScope"`
	CustomDeptIDs []int64 `json:"customDeptIds"`
}

// NewRoleDataScopeAll 创建"全部数据"权限
func NewRoleDataScopeAll(roleCode string) RoleDataScope {
	return RoleDataScope{RoleCode: roleCode, DataScope: 1, CustomDeptIDs: nil}
}

// NewRoleDataScopeDeptAndSub 创建"部门及子部门"权限
func NewRoleDataScopeDeptAndSub(roleCode string) RoleDataScope {
	return RoleDataScope{RoleCode: roleCode, DataScope: 2, CustomDeptIDs: nil}
}

// NewRoleDataScopeDept 创建"本部门"权限
func NewRoleDataScopeDept(roleCode string) RoleDataScope {
	return RoleDataScope{RoleCode: roleCode, DataScope: 3, CustomDeptIDs: nil}
}

// NewRoleDataScopeSelf 创建"本人"权限
func NewRoleDataScopeSelf(roleCode string) RoleDataScope {
	return RoleDataScope{RoleCode: roleCode, DataScope: 4, CustomDeptIDs: nil}
}

// NewRoleDataScopeCustom 创建"自定义部门"权限
func NewRoleDataScopeCustom(roleCode string, deptIDs []int64) RoleDataScope {
	return RoleDataScope{RoleCode: roleCode, DataScope: 5, CustomDeptIDs: deptIDs}
}

// UserPermissionsVO 用户权限信息
type UserPermissionsVO struct {
	UserID     types.BigInt    `json:"userId"`
	Roles      []string        `json:"roles"`
	Perms      []string        `json:"perms"`
	DataScopes []RoleDataScope `json:"dataScopes"`
	DeptID     types.BigInt    `json:"deptId"`
}
