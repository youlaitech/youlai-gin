package model

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"` // 用户名
	Password string `json:"password" binding:"required" example:"123456"` // 密码
}

// CaptchaVO 验证码信息
type CaptchaVO struct {
	CaptchaKey    string `json:"captchaKey"`    // 验证码缓存 Key
	CaptchaBase64 string `json:"captchaBase64"` // 验证码图片 Base64 字符串
}

// SmsLoginRequest 短信验证码登录请求
type SmsLoginRequest struct {
	Mobile string `json:"mobile" binding:"required" example:"18812345678"` // 手机号
	Code   string `json:"code" binding:"required" example:"1234"`          // 验证码
}

// WxMiniAppCodeLoginRequest 微信小程序Code登录请求
type WxMiniAppCodeLoginRequest struct {
	Code string `json:"code" binding:"required"` // 微信小程序登录时获取的code
}

// WxMiniAppPhoneLoginRequest 微信小程序手机号登录请求
type WxMiniAppPhoneLoginRequest struct {
	Code          string `json:"code" binding:"required"`     // 微信小程序登录时获取的code
	EncryptedData string `json:"encryptedData"`                // 包括敏感数据在内的完整用户信息的加密数据
	IV            string `json:"iv"`                           // 加密算法的初始向量
}

// LoginResponse 登录响应（使用 pkg/auth 的 AuthenticationToken）
// 为了保持一致性，这里直接引用
