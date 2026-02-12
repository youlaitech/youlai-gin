package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	redisClient "youlai-gin/pkg/redis"
)

// Redis Key 常量
const (
	AccessTokenUserPrefix     = "auth:token:access:"
	RefreshTokenUserPrefix    = "auth:token:refresh:"
	UserAccessTokenPrefix     = "auth:user:access:"
	UserRefreshTokenPrefix    = "auth:user:refresh:"
	BlacklistTokenPrefix      = "auth:token:blacklist:"
	UserTokenValidAfterPrefix = "auth:user:token_valid_after:"
)

// tokenValidAfter 默认过期时间（7天），避免Redis内存泄漏
const TokenValidAfterTTLSeconds = 7 * 24 * 60 * 60

// RedisTokenConfig Redis Token 配置
type RedisTokenConfig struct {
	AccessTokenTTL  int  // 访问令牌过期时间（秒）
	RefreshTokenTTL int  // 刷新令牌过期时间（秒）
	AllowMultiLogin bool // 是否允许多设备登录
}

// RedisTokenManager Redis Token 管理器
// 实现基于Redis的有状态认证，支持：
// - Access Token + Refresh Token 双令牌机制
// - 单设备/多设备登录控制
// - 用户级会话失效
// - 在线用户管理
type RedisTokenManager struct {
	config *RedisTokenConfig
}

// NewRedisTokenManager 创建 Redis Token 管理器
func NewRedisTokenManager(config *RedisTokenConfig) *RedisTokenManager {
	return &RedisTokenManager{config: config}
}

// GenerateToken 生成认证 Token
func (m *RedisTokenManager) GenerateToken(user *UserDetails) (*AuthenticationToken, error) {
	accessToken := uuid.New().String()
	refreshToken := uuid.New().String()

	userSession := &UserSession{
		UserID:     user.UserID,
		Username:   user.Username,
		DeptID:     user.DeptID,
		DataScopes: user.DataScopes,
		Roles:      user.Roles,
	}

	ctx := context.Background()

	// 1. 存储访问令牌 -> 用户会话信息
	if err := m.storeUserSession(ctx, accessToken, userSession, m.config.AccessTokenTTL); err != nil {
		return nil, err
	}

	// 2. 存储刷新令牌 -> 用户会话信息
	refreshKey := RefreshTokenUserPrefix + refreshToken
	if err := m.storeUserSession(ctx, refreshKey, userSession, m.config.RefreshTokenTTL); err != nil {
		return nil, err
	}

	// 3. 存储用户ID -> 刷新令牌
	userRefreshKey := fmt.Sprintf("%s%d", UserRefreshTokenPrefix, user.UserID)
	if err := m.setWithTTL(ctx, userRefreshKey, refreshToken, m.config.RefreshTokenTTL); err != nil {
		return nil, err
	}

	// 4. 单设备登录控制
	if err := m.handleSingleDeviceLogin(ctx, user.UserID, accessToken); err != nil {
		return nil, err
	}

	return &AuthenticationToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    m.config.AccessTokenTTL,
	}, nil
}

// ParseToken 解析 Token 获取用户信息
func (m *RedisTokenManager) ParseToken(token string) (*UserDetails, error) {
	ctx := context.Background()
	key := AccessTokenUserPrefix + token

	data, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.New("token not found or expired")
	}

	var userSession UserSession
	if err := json.Unmarshal([]byte(data), &userSession); err != nil {
		return nil, err
	}

	return userSession.ToUserDetails(), nil
}

// ValidateToken 校验 Token 是否有效
func (m *RedisTokenManager) ValidateToken(token string) bool {
	ctx := context.Background()
	key := AccessTokenUserPrefix + token
	exists, err := redisClient.Client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// ValidateRefreshToken 校验刷新 Token 是否有效
func (m *RedisTokenManager) ValidateRefreshToken(refreshToken string) bool {
	ctx := context.Background()
	key := RefreshTokenUserPrefix + refreshToken
	exists, err := redisClient.Client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// RefreshToken 刷新 Token
func (m *RedisTokenManager) RefreshToken(refreshToken string) (*AuthenticationToken, error) {
	if !m.ValidateRefreshToken(refreshToken) {
		return nil, errors.New("invalid refresh token")
	}

	ctx := context.Background()
	refreshKey := RefreshTokenUserPrefix + refreshToken

	data, err := redisClient.Client.Get(ctx, refreshKey).Result()
	if err != nil {
		return nil, errors.New("refresh token expired")
	}

	var userSession UserSession
	if err := json.Unmarshal([]byte(data), &userSession); err != nil {
		return nil, err
	}

	// 删除旧的访问令牌
	userAccessKey := fmt.Sprintf("%s%d", UserAccessTokenPrefix, userSession.UserID)
	oldAccessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
	if err == nil {
		oldKey := AccessTokenUserPrefix + oldAccessToken
		redisClient.Client.Del(ctx, oldKey)
	}

	// 生成新访问令牌
	newAccessToken := uuid.New().String()
	if err := m.storeUserSession(ctx, newAccessToken, &userSession, m.config.AccessTokenTTL); err != nil {
		return nil, err
	}

	// 更新用户ID -> 访问令牌映射
	if err := m.setWithTTL(ctx, userAccessKey, newAccessToken, m.config.AccessTokenTTL); err != nil {
		return nil, err
	}

	return &AuthenticationToken{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    m.config.AccessTokenTTL,
	}, nil
}

// InvalidateToken 令 Token 失效
func (m *RedisTokenManager) InvalidateToken(token string) error {
	ctx := context.Background()
	key := AccessTokenUserPrefix + token

	data, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		return nil // Token 不存在或已过期
	}

	var userSession UserSession
	if err := json.Unmarshal([]byte(data), &userSession); err != nil {
		return err
	}

	return m.InvalidateUserSessions(userSession.UserID)
}

// InvalidateUserSessions 使指定用户的所有会话失效
// 适用场景：用户修改密码、管理员强制下线、账号封禁等
func (m *RedisTokenManager) InvalidateUserSessions(userID int64) error {
	ctx := context.Background()

	// 1. 删除访问令牌
	userAccessKey := fmt.Sprintf("%s%d", UserAccessTokenPrefix, userID)
	accessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
	if err == nil {
		accessKey := AccessTokenUserPrefix + accessToken
		redisClient.Client.Del(ctx, accessKey)
	}
	redisClient.Client.Del(ctx, userAccessKey)

	// 2. 删除刷新令牌
	userRefreshKey := fmt.Sprintf("%s%d", UserRefreshTokenPrefix, userID)
	refreshToken, err := redisClient.Client.Get(ctx, userRefreshKey).Result()
	if err == nil {
		refreshKey := RefreshTokenUserPrefix + refreshToken
		redisClient.Client.Del(ctx, refreshKey)
	}
	redisClient.Client.Del(ctx, userRefreshKey)

	return nil
}

// SetTokenValidAfter 设置用户 Token 生效时间点
// 用于JWT模式下的会话失效控制，设置TTL防止Redis内存泄漏
func (m *RedisTokenManager) SetTokenValidAfter(userID int64) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", UserTokenValidAfterPrefix, userID)
	now := time.Now().Unix()
	return redisClient.Client.Set(ctx, key, now, time.Duration(TokenValidAfterTTLSeconds)*time.Second).Err()
}

// GetTokenValidAfter 获取用户 Token 生效时间点
func (m *RedisTokenManager) GetTokenValidAfter(userID int64) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", UserTokenValidAfterPrefix, userID)
	result, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	var value int64
	fmt.Sscanf(result, "%d", &value)
	return value, nil
}

// storeUserSession 存储用户会话信息
func (m *RedisTokenManager) storeUserSession(ctx context.Context, keyOrToken string, session *UserSession, ttl int) error {
	var key string
	if len(keyOrToken) == 36 { // UUID 格式
		key = AccessTokenUserPrefix + keyOrToken
	} else {
		key = keyOrToken
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return m.setWithTTL(ctx, key, string(data), ttl)
}

// handleSingleDeviceLogin 处理单设备登录控制
func (m *RedisTokenManager) handleSingleDeviceLogin(ctx context.Context, userID int64, newAccessToken string) error {
	userAccessKey := fmt.Sprintf("%s%d", UserAccessTokenPrefix, userID)

	// 单设备登录：删除旧令牌
	if !m.config.AllowMultiLogin {
		oldAccessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
		if err == nil {
			oldKey := AccessTokenUserPrefix + oldAccessToken
			redisClient.Client.Del(ctx, oldKey)
		}
	}

	// 存储新令牌映射
	return m.setWithTTL(ctx, userAccessKey, newAccessToken, m.config.AccessTokenTTL)
}

// setWithTTL 设置带过期时间的值
func (m *RedisTokenManager) setWithTTL(ctx context.Context, key, value string, ttl int) error {
	if ttl == -1 {
		return redisClient.Client.Set(ctx, key, value, 0).Err()
	}
	return redisClient.Client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}
