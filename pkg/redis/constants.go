package redis

// Redis Key 常量定义（对应 Java RedisConstants）
const (
	// 认证相关
	AccessTokenUserPrefix  = "auth:access_token:"         // 访问令牌 -> 用户信息
	RefreshTokenUserPrefix = "auth:refresh_token:"        // 刷新令牌 -> 用户信息
	UserAccessTokenPrefix  = "auth:user:access_token:"    // 用户ID -> 访问令牌
	UserRefreshTokenPrefix = "auth:user:refresh_token:"   // 用户ID -> 刷新令牌
	BlacklistTokenPrefix   = "auth:blacklist:token:"      // Token 黑名单
	UserSecurityVersion    = "auth:user:security_version:" // 用户安全版本号
)
