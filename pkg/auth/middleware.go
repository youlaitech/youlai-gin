package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"youlai-gin/pkg/response"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
)

// Middleware 认证中间件
func Middleware(tokenManager TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 中获取 Token
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.TokenInvalid(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 验证 Bearer 前缀
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.TokenInvalid(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		// 提取 Token
		token := strings.TrimPrefix(authHeader, BearerPrefix)

		// 校验 Token
		if !tokenManager.ValidateToken(token) {
			response.TokenInvalid(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 解析用户信息
		user, err := tokenManager.ParseToken(token)
		if err != nil {
			response.TokenInvalid(c, "解析认证令牌失败")
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
