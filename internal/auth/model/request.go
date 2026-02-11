package model

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"` // 用户名
	Password string `json:"password" binding:"required" example:"123456"` // 密码
}

// SmsLoginRequest 短信验证码登录请求
type SmsLoginRequest struct {
	Mobile string `json:"mobile" binding:"required" example:"18812345678"` // 手机号
	Code   string `json:"code" binding:"required" example:"1234"`          // 验证码
}
