package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/pkg/errs"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
)

// 认证白名单：命中则跳过 JWT 校验
var authWhitelist = map[string]bool{
	"/api/v1/statistics/visits/trend":    true,
	"/api/v1/statistics/visits/overview": true,
	"/api/v1/auth/qr-code/generate":      true,
	"/api/v1/auth/qr-code/status":        true,
	"/api/v1/auth/qr-code/login":         true,
	"/api/v1/realtime/ws":                true,
}

// 匿名前缀：路径含参数段，精确匹配覆盖不到（如公开表单 /api/v1/forms/public/{formKey}/render）
var anonymousPrefixes = []string{
	"/api/v1/forms/public/",
}

// isAnonymous 路径是否免认证
func isAnonymous(path string) bool {
	if authWhitelist[path] {
		return true
	}
	for _, prefix := range anonymousPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// Middleware 认证中间件
func Middleware(tokenManager TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单路径跳过认证
		if isAnonymous(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 从 Header 中获取 Token
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Error(errs.TokenInvalid())
			c.Abort()
			return
		}

		// 验证 Bearer 前缀
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Error(errs.TokenInvalid())
			c.Abort()
			return
		}

		// 提取 Token
		token := strings.TrimPrefix(authHeader, BearerPrefix)

		// 校验 Token
		if !tokenManager.ValidateToken(token) {
			c.Error(errs.TokenInvalid())
			c.Abort()
			return
		}

		// 解析用户信息
		user, err := tokenManager.ParseToken(token)
		if err != nil {
			c.Error(errs.TokenInvalid())
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(UserContextKey, user)
		c.Next()
	}
}

// GetCurrentUser 从上下文获取当前用户
func GetCurrentUser(c *gin.Context) (*UserDetails, bool) {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	userDetails, ok := user.(*UserDetails)
	return userDetails, ok
}
