package model

// WechatSilentLoginRequest 静默登录请求
type WechatSilentLoginRequest struct {
	Code string `json:"code" binding:"required" example:"xxx"` // 微信登录code
}

// WechatPhoneLoginRequest 手机号快捷登录请求
type WechatPhoneLoginRequest struct {
	LoginCode string `json:"loginCode" binding:"required" example:"xxx"` // 微信登录code
	PhoneCode string `json:"phoneCode" binding:"required" example:"xxx"` // 微信手机号code
}

// WechatBindMobileRequest 绑定手机号请求
type WechatBindMobileRequest struct {
	OpenID string `json:"openId" binding:"required" example:"xxx"` // 微信openid
	Mobile string `json:"mobile" binding:"required" example:"18812345678"` // 手机号
	SmsCode string `json:"smsCode" binding:"required" example:"1234"` // 短信验证码
}

// WechatMiniappLoginResult 微信小程序登录结果
type WechatMiniappLoginResult struct {
	NeedBindMobile bool   `json:"needBindMobile"`         // 是否需要绑定手机号
	AccessToken    string `json:"accessToken,omitempty"`  // 访问令牌
	RefreshToken   string `json:"refreshToken,omitempty"` // 刷新令牌
	ExpiresIn      int64  `json:"expiresIn,omitempty"`    // 令牌过期时间(秒)
	TokenType      string `json:"tokenType,omitempty"`    // 令牌类型
	OpenID         string `json:"openId,omitempty"`       // 微信openid（绑定手机号时使用）
}
