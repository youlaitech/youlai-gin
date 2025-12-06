package auth

// TokenManager Token 管理器接口
// 用于生成、解析、校验、刷新 Token
type TokenManager interface {
	// GenerateToken 生成认证 Token
	GenerateToken(user *UserDetails) (*AuthenticationToken, error)

	// ParseToken 解析 Token 获取用户信息
	ParseToken(token string) (*UserDetails, error)

	// ValidateToken 校验 Token 是否有效
	ValidateToken(token string) bool

	// ValidateRefreshToken 校验刷新 Token 是否有效
	ValidateRefreshToken(refreshToken string) bool

	// RefreshToken 刷新 Token
	RefreshToken(refreshToken string) (*AuthenticationToken, error)

	// InvalidateToken 令 Token 失效
	InvalidateToken(token string) error

	// InvalidateUserSessions 使指定用户的所有会话失效
	InvalidateUserSessions(userID int64) error
}
