package model

// CaptchaVO 验证码信息
type CaptchaVO struct {
	CaptchaKey    string `json:"captchaKey"`    // 验证码缓存 Key
	CaptchaBase64 string `json:"captchaBase64"` // 验证码图片 Base64 字符串
}
