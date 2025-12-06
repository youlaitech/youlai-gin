package model

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"` // 用户名
	Password string `json:"password" binding:"required" example:"123456"` // 密码
}

// LoginResponse 登录响应（使用 pkg/auth 的 AuthenticationToken）
// 为了保持一致性，这里直接引用
