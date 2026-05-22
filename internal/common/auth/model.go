package auth

import (
	permModel "youlai-gin/internal/common/permission/model"
	"youlai-gin/pkg/types"
)

// AuthenticationToken 认证令牌响应
type AuthenticationToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"` // Bearer
	ExpiresIn    int    `json:"expiresIn"` // 过期时间（秒）
}

// UserDetails 用户详情
type UserDetails struct {
	UserID     int64                   `json:"userId"`
	Username   string                  `json:"username"`
	DeptID     types.BigInt            `json:"deptId"`
	DataScopes []permModel.RoleDataScope `json:"dataScopes"` // 数据权限列表（支持多角色）
	Roles      []string                `json:"roles"`        // 角色列表
}

// UserSession Redis-Token 模式下的用户会话快照
type UserSession struct {
	UserID     int64                    `json:"userId"`
	Username   string                   `json:"username"`
	DeptID     types.BigInt             `json:"deptId"`
	DataScopes []permModel.RoleDataScope `json:"dataScopes"` // 数据权限列表
	Roles      []string                 `json:"roles"`       // 角色权限集合
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