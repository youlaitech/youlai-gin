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

// Middleware 认证中间件
func Middleware(tokenManager TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单路径跳过认证
		if c.Request.URL.Path == "/api/v1/statistics/visits/trend" || c.Request.URL.Path == "/api/v1/statistics/visits/overview" {
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
