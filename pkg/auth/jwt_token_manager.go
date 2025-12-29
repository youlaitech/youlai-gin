package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	redisClient "youlai-gin/pkg/redis"
	"youlai-gin/pkg/types"
)

// JwtConfig JWT 配置
type JwtConfig struct {
	SecretKey              string // 密钥
	AccessTokenTTL         int    // 访问令牌过期时间（秒）
	RefreshTokenTTL        int    // 刷新令牌过期时间（秒）
	EnableSecurityVersion  bool   // 是否启用安全版本号
}

// JwtTokenManager JWT Token 管理器
type JwtTokenManager struct {
	config *JwtConfig
}

// CustomClaims 自定义 Claims
type CustomClaims struct {
	UserID          int64       `json:"userId"`
	Username        string      `json:"username"`
	DeptID          types.BigInt `json:"deptId"`
	DataScope       int         `json:"dataScope"`
	Roles           []string    `json:"roles"`
	IsRefreshToken  bool        `json:"isRefreshToken"`  // 是否为刷新令牌
	SecurityVersion int         `json:"securityVersion"` // 安全版本号
	jwt.RegisteredClaims
}

// NewJwtTokenManager 创建 JWT Token 管理器
func NewJwtTokenManager(config *JwtConfig) *JwtTokenManager {
	return &JwtTokenManager{config: config}
}

// GenerateToken 生成认证 Token
func (m *JwtTokenManager) GenerateToken(user *UserDetails) (*AuthenticationToken, error) {
	accessToken, err := m.generateToken(user, m.config.AccessTokenTTL, false)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generateToken(user, m.config.RefreshTokenTTL, true)
	if err != nil {
		return nil, err
	}

	return &AuthenticationToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    m.config.AccessTokenTTL,
	}, nil
}

// generateToken 生成 JWT Token
func (m *JwtTokenManager) generateToken(user *UserDetails, ttl int, isRefreshToken bool) (string, error) {
	now := time.Now()
	var exp time.Time
	if ttl != -1 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	// 获取安全版本号
	securityVersion := 0
	if m.config.EnableSecurityVersion {
		ctx := context.Background()
		key := fmt.Sprintf("%s%d", redisClient.UserSecurityVersion, user.UserID)
		val, err := redisClient.Client.Get(ctx, key).Int()
		if err == nil {
			securityVersion = val
		}
	}

	claims := CustomClaims{
		UserID:          user.UserID,
		Username:        user.Username,
		DeptID:          user.DeptID,
		DataScope:       user.DataScope,
		Roles:           user.Roles,
		IsRefreshToken:  isRefreshToken,
		SecurityVersion: securityVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// ParseToken 解析 Token 获取用户信息
func (m *JwtTokenManager) ParseToken(tokenString string) (*UserDetails, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return &UserDetails{
			UserID:    claims.UserID,
			Username:  claims.Username,
			DeptID:    claims.DeptID,
			DataScope: claims.DataScope,
			Roles:     claims.Roles,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateToken 校验 Token 是否有效
func (m *JwtTokenManager) ValidateToken(tokenString string) bool {
	return m.validateToken(tokenString, false)
}

// ValidateRefreshToken 校验刷新 Token 是否有效
func (m *JwtTokenManager) ValidateRefreshToken(tokenString string) bool {
	return m.validateToken(tokenString, true)
}

// validateToken 校验 Token
func (m *JwtTokenManager) validateToken(tokenString string, isRefreshToken bool) bool {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return false
	}

	// 校验刷新令牌类型
	if isRefreshToken && !claims.IsRefreshToken {
		return false
	}

	ctx := context.Background()

	// 校验安全版本号
	if m.config.EnableSecurityVersion {
		key := fmt.Sprintf("%s%d", redisClient.UserSecurityVersion, claims.UserID)
		currentVersion, err := redisClient.Client.Get(ctx, key).Int()
		if err == nil && claims.SecurityVersion < currentVersion {
			return false // 版本号过期
		}
	}

	// 校验黑名单
	blacklistKey := fmt.Sprintf("%s%s", redisClient.BlacklistTokenPrefix, claims.ID)
	exists, err := redisClient.Client.Exists(ctx, blacklistKey).Result()
	if err == nil && exists > 0 {
		return false // 在黑名单中
	}

	return true
}

// RefreshToken 刷新 Token
func (m *JwtTokenManager) RefreshToken(refreshToken string) (*AuthenticationToken, error) {
	if !m.ValidateRefreshToken(refreshToken) {
		return nil, errors.New("invalid refresh token")
	}

	user, err := m.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 生成新的访问令牌
	accessToken, err := m.generateToken(user, m.config.AccessTokenTTL, false)
	if err != nil {
		return nil, err
	}

	return &AuthenticationToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    m.config.AccessTokenTTL,
	}, nil
}

// InvalidateToken 令 Token 失效（加入黑名单）
func (m *JwtTokenManager) InvalidateToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// 计算剩余过期时间
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL <= 0 {
		return nil // 已过期，无需加入黑名单
	}

	ctx := context.Background()
	blacklistKey := fmt.Sprintf("%s%s", redisClient.BlacklistTokenPrefix, claims.ID)
	return redisClient.Client.Set(ctx, blacklistKey, true, remainingTTL).Err()
}

// InvalidateUserSessions 使指定用户的所有会话失效
func (m *JwtTokenManager) InvalidateUserSessions(userID int64) error {
	if !m.config.EnableSecurityVersion {
		return errors.New("security version not enabled")
	}

	ctx := context.Background()
	key := fmt.Sprintf("%s%d", redisClient.UserSecurityVersion, userID)
	return redisClient.Client.Incr(ctx, key).Err()
}
