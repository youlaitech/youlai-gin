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

// RedisTokenConfig Redis Token 配置
type RedisTokenConfig struct {
	AccessTokenTTL   int  // 访问令牌过期时间（秒）
	RefreshTokenTTL  int  // 刷新令牌过期时间（秒）
	AllowMultiLogin  bool // 是否允许多设备登录
}

// RedisTokenManager Redis Token 管理器
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

	onlineUser := &OnlineUser{
		UserID:    user.UserID,
		Username:  user.Username,
		DeptID:    user.DeptID,
		DataScope: user.DataScope,
		Roles:     user.Roles,
	}

	ctx := context.Background()

	// 1. 存储访问令牌 -> 用户信息
	if err := m.storeOnlineUser(ctx, accessToken, onlineUser, m.config.AccessTokenTTL); err != nil {
		return nil, err
	}

	// 2. 存储刷新令牌 -> 用户信息
	refreshKey := fmt.Sprintf("%s%s", redisClient.RefreshTokenUserPrefix, refreshToken)
	if err := m.storeOnlineUser(ctx, refreshKey, onlineUser, m.config.RefreshTokenTTL); err != nil {
		return nil, err
	}

	// 3. 存储用户ID -> 刷新令牌
	userRefreshKey := fmt.Sprintf("%s%d", redisClient.UserRefreshTokenPrefix, user.UserID)
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
	key := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, token)

	data, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.New("token not found or expired")
	}

	var onlineUser OnlineUser
	if err := json.Unmarshal([]byte(data), &onlineUser); err != nil {
		return nil, err
	}

	return &UserDetails{
		UserID:    onlineUser.UserID,
		Username:  onlineUser.Username,
		DeptID:    onlineUser.DeptID,
		DataScope: onlineUser.DataScope,
		Roles:     onlineUser.Roles,
	}, nil
}

// ValidateToken 校验 Token 是否有效
func (m *RedisTokenManager) ValidateToken(token string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, token)
	exists, err := redisClient.Client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// ValidateRefreshToken 校验刷新 Token 是否有效
func (m *RedisTokenManager) ValidateRefreshToken(refreshToken string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s", redisClient.RefreshTokenUserPrefix, refreshToken)
	exists, err := redisClient.Client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// RefreshToken 刷新 Token
func (m *RedisTokenManager) RefreshToken(refreshToken string) (*AuthenticationToken, error) {
	if !m.ValidateRefreshToken(refreshToken) {
		return nil, errors.New("invalid refresh token")
	}

	ctx := context.Background()
	refreshKey := fmt.Sprintf("%s%s", redisClient.RefreshTokenUserPrefix, refreshToken)

	data, err := redisClient.Client.Get(ctx, refreshKey).Result()
	if err != nil {
		return nil, errors.New("refresh token expired")
	}

	var onlineUser OnlineUser
	if err := json.Unmarshal([]byte(data), &onlineUser); err != nil {
		return nil, err
	}

	// 删除旧的访问令牌
	userAccessKey := fmt.Sprintf("%s%d", redisClient.UserAccessTokenPrefix, onlineUser.UserID)
	oldAccessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
	if err == nil {
		oldKey := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, oldAccessToken)
		redisClient.Client.Del(ctx, oldKey)
	}

	// 生成新访问令牌
	newAccessToken := uuid.New().String()
	if err := m.storeOnlineUser(ctx, newAccessToken, &onlineUser, m.config.AccessTokenTTL); err != nil {
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
	key := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, token)

	data, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		return nil // Token 不存在或已过期
	}

	var onlineUser OnlineUser
	if err := json.Unmarshal([]byte(data), &onlineUser); err != nil {
		return err
	}

	return m.InvalidateUserSessions(onlineUser.UserID)
}

// InvalidateUserSessions 使指定用户的所有会话失效
func (m *RedisTokenManager) InvalidateUserSessions(userID int64) error {
	ctx := context.Background()

	// 1. 删除访问令牌
	userAccessKey := fmt.Sprintf("%s%d", redisClient.UserAccessTokenPrefix, userID)
	accessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
	if err == nil {
		accessKey := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, accessToken)
		redisClient.Client.Del(ctx, accessKey)
	}
	redisClient.Client.Del(ctx, userAccessKey)

	// 2. 删除刷新令牌
	userRefreshKey := fmt.Sprintf("%s%d", redisClient.UserRefreshTokenPrefix, userID)
	refreshToken, err := redisClient.Client.Get(ctx, userRefreshKey).Result()
	if err == nil {
		refreshKey := fmt.Sprintf("%s%s", redisClient.RefreshTokenUserPrefix, refreshToken)
		redisClient.Client.Del(ctx, refreshKey)
	}
	redisClient.Client.Del(ctx, userRefreshKey)

	return nil
}

// storeOnlineUser 存储在线用户信息
func (m *RedisTokenManager) storeOnlineUser(ctx context.Context, keyOrToken string, user *OnlineUser, ttl int) error {
	var key string
	if len(keyOrToken) == 36 { // UUID 格式
		key = fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, keyOrToken)
	} else {
		key = keyOrToken
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return m.setWithTTL(ctx, key, string(data), ttl)
}

// handleSingleDeviceLogin 处理单设备登录控制
func (m *RedisTokenManager) handleSingleDeviceLogin(ctx context.Context, userID int64, newAccessToken string) error {
	userAccessKey := fmt.Sprintf("%s%d", redisClient.UserAccessTokenPrefix, userID)

	// 单设备登录：删除旧令牌
	if !m.config.AllowMultiLogin {
		oldAccessToken, err := redisClient.Client.Get(ctx, userAccessKey).Result()
		if err == nil {
			oldKey := fmt.Sprintf("%s%s", redisClient.AccessTokenUserPrefix, oldAccessToken)
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
