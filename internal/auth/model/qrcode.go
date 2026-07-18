package model

// 扫码登录状态机：WAITING → SCANNED → CONFIRMED → LOGGED_IN，外加 CANCELED、EXPIRED
const (
	QrCodeStatusWaiting   = "WAITING"   // 票据已建，待 APP 扫码
	QrCodeStatusScanned   = "SCANNED"   // APP 已扫码，待用户确认
	QrCodeStatusConfirmed = "CONFIRMED" // 用户已在 APP 确认
	QrCodeStatusLoggedIn  = "LOGGED_IN" // PC 已用票据换令牌，票据作废
	QrCodeStatusCanceled  = "CANCELED"  // APP 端用户取消
	QrCodeStatusExpired   = "EXPIRED"   // Redis TTL 自动过期（内存中通常不出现）
)

// QrCodeLoginContext 扫码登录票据上下文，以 JSON 形式存入 Redis
type QrCodeLoginContext struct {
	Ticket      string `json:"ticket"`      // 票据号（UUID 去连字符）
	Status      string `json:"status"`      // 当前状态，取值见上方常量
	UserID      int64  `json:"userId"`      // 扫码用户 ID，未扫码时为 0
	Nickname    string `json:"nickname"`    // 扫码用户昵称
	Avatar      string `json:"avatar"`      // 扫码用户头像
	CreatedAt   int64  `json:"createdAt"`   // 票据创建时间（毫秒）
	ScannedAt   int64  `json:"scannedAt"`   // 扫码时间（毫秒）
	ConfirmedAt int64  `json:"confirmedAt"` // 确认时间（毫秒）
	ClientIP    string `json:"clientIp"`    // 生成票据时的客户端 IP
}

// QrCodeGenerateVO 生成票据接口返回
type QrCodeGenerateVO struct {
	Ticket        string `json:"ticket"`        // 票据号
	ExpireSeconds int    `json:"expireSeconds"` // 票据有效期（秒）
}

// QrCodeStatusVO 票据状态接口返回；WAITING 阶段 nickname/avatar 为 null，不泄露用户信息
type QrCodeStatusVO struct {
	Ticket        string  `json:"ticket"`
	Status        string  `json:"status"`
	Nickname      *string `json:"nickname"`
	Avatar        *string `json:"avatar"`
	ExpireSeconds int     `json:"expireSeconds"`
}

// QrCodeTicketForm 扫码操作请求体，携带票据号
type QrCodeTicketForm struct {
	Ticket string `json:"ticket" binding:"required"` // 票据号
}
