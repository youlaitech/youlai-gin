package model

// Log 操作日志实体（对应 sys_log 表）
type Log struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Module       string `gorm:"column:module;size:50" json:"module"`             // 操作模块
	Operation    string `gorm:"column:operation;size:50" json:"operation"`       // 操作类型
	Method       string `gorm:"column:method;size:10" json:"method"`             // 请求方法
	Path         string `gorm:"column:path;size:255" json:"path"`                // 请求路径
	UserID       int64  `gorm:"column:user_id" json:"userId"`                    // 操作用户ID
	Username     string `gorm:"column:username;size:50" json:"username"`         // 操作用户名
	IP           string `gorm:"column:ip;size:50" json:"ip"`                     // IP地址
	UserAgent    string `gorm:"column:user_agent;size:500" json:"userAgent"`     // User-Agent
	RequestBody  string `gorm:"column:request_body;type:text" json:"requestBody"` // 请求体
	ResponseBody string `gorm:"column:response_body;type:text" json:"responseBody"` // 响应体
	Status       int    `gorm:"column:status" json:"status"`                     // 响应状态码
	Duration     int64  `gorm:"column:duration" json:"duration"`                 // 执行时长（毫秒）
	ErrorMsg     string `gorm:"column:error_msg;size:500" json:"errorMsg"`       // 错误信息
	CreateTime   string `gorm:"column:create_time;autoCreateTime" json:"createTime"` // 创建时间
}

func (Log) TableName() string {
	return "sys_log"
}
