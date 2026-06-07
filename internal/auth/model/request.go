package model

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required" example:"admin"`           // 用户名
	Password    string `json:"password" binding:"required" example:"123456"`          // 密码
	CaptchaID   string `json:"captchaId" binding:"required" example:"captcha_id_123"` // 验证码缓存 ID
	CaptchaCode string `json:"captchaCode" binding:"required" example:"1234"`         // 验证码
}

// SmsLoginRequest 短信验证码登录请求
type SmsLoginRequest struct {
	Mobile string `json:"mobile" binding:"required" example:"18888888888"` // 手机号
	Code   string `json:"code" binding:"required" example:"1234"`          // 验证码
}
