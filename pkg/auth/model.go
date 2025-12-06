package auth

// AuthenticationToken 认证令牌响应
type AuthenticationToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"` // Bearer
	ExpiresIn    int    `json:"expiresIn"` // 过期时间（秒）
}

// UserDetails 用户详情
type UserDetails struct {
	UserID    int64    `json:"userId"`
	Username  string   `json:"username"`
	DeptID    int64    `json:"deptId"`
	DataScope int      `json:"dataScope"` // 数据权限范围
	Roles     []string `json:"roles"`     // 角色列表
}

// OnlineUser Redis 存储的在线用户信息
type OnlineUser struct {
	UserID    int64    `json:"userId"`
	Username  string   `json:"username"`
	DeptID    int64    `json:"deptId"`
	DataScope int      `json:"dataScope"`
	Roles     []string `json:"roles"`
}
