package auth

import "youlai-gin/pkg/types"

// AuthenticationToken 认证令牌响应
type AuthenticationToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"` // Bearer
	ExpiresIn    int    `json:"expiresIn"` // 过期时间（秒）
}

// RoleDataScope 角色数据权限信息
// 用于存储单个角色的数据权限范围信息，支持多角色数据权限合并（并集策略）
type RoleDataScope struct {
	RoleCode      string  `json:"roleCode"`      // 角色编码
	DataScope     int     `json:"dataScope"`     // 数据权限范围值：1-所有数据 2-部门及子部门 3-本部门 4-本人 5-自定义
	CustomDeptIDs []int64 `json:"customDeptIds"` // 自定义部门ID列表（仅当 dataScope=5 时有效）
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

// UserDetails 用户详情
type UserDetails struct {
	UserID     int64          `json:"userId"`
	Username   string         `json:"username"`
	DeptID     types.BigInt   `json:"deptId"`
	DataScopes []RoleDataScope `json:"dataScopes"` // 数据权限列表（支持多角色）
	Roles      []string       `json:"roles"`      // 角色列表
}

// UserSession 用户会话信息
// 存储在Token中的用户会话快照，包含用户身份、数据权限和角色权限信息。
// 用于Redis-Token模式下的会话管理，支持在线用户查询和会话控制。
type UserSession struct {
	UserID     int64           `json:"userId"`
	Username   string          `json:"username"`
	DeptID     types.BigInt    `json:"deptId"`
	DataScopes []RoleDataScope `json:"dataScopes"` // 数据权限列表
	Roles      []string        `json:"roles"`      // 角色权限集合
}

// ToUserDetails 转换为 UserDetails
func (s *UserSession) ToUserDetails() *UserDetails {
	return &UserDetails{
		UserID:     s.UserID,
		Username:   s.Username,
		DeptID:     s.DeptID,
		DataScopes: s.DataScopes,
		Roles:      s.Roles,
	}
}
