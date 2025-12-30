package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"youlai-gin/pkg/database"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/logger"
)

// OperationLog 操作日志实体（对应 sys_log 表）
type OperationLogEntity struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Module        string `gorm:"column:module;size:50" json:"module"`               // 操作模块
	Operation     string `gorm:"column:operation;size:50" json:"operation"`         // 操作类型
	Method        string `gorm:"column:method;size:10" json:"method"`               // 请求方法
	Path          string `gorm:"column:path;size:255" json:"path"`                  // 请求路径
	UserID        int64  `gorm:"column:user_id" json:"userId"`                      // 操作用户ID
	Username      string `gorm:"column:username;size:50" json:"username"`           // 操作用户名
	IP            string `gorm:"column:ip;size:50" json:"ip"`                       // IP地址
	UserAgent     string `gorm:"column:user_agent;size:500" json:"userAgent"`       // User-Agent
	RequestBody   string `gorm:"column:request_body;type:text" json:"requestBody"`  // 请求体
	ResponseBody  string `gorm:"column:response_body;type:text" json:"responseBody"`// 响应体
	Status        int    `gorm:"column:status" json:"status"`                       // 响应状态码
	Duration      int64  `gorm:"column:duration" json:"duration"`                   // 执行时长（毫秒）
	ErrorMsg      string `gorm:"column:error_msg;size:500" json:"errorMsg"`         // 错误信息
	CreateTime    string `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (OperationLogEntity) TableName() string {
	return "sys_log"
}

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Module          string // 操作模块
	Operation       string // 操作类型
	SaveRequestBody bool   // 是否保存请求体
	SaveResponse    bool   // 是否保存响应体
	MaxBodySize     int    // 最大请求体大小（字节），0 表示不限制
}

// DefaultOperationLogConfig 默认操作日志配置
var DefaultOperationLogConfig = OperationLogConfig{
	SaveRequestBody: true,
	SaveResponse:    false, // 默认不保存响应（数据量大）
	MaxBodySize:     10240, // 默认最大 10KB
}

// OperationLog 操作日志中间件（装饰器）
// 用法：router.POST("/users", middleware.OperationLog("用户管理", "新增用户"), handler.CreateUser)
func OperationLog(module, operation string) gin.HandlerFunc {
	return OperationLogWithConfig(OperationLogConfig{
		Module:          module,
		Operation:       operation,
		SaveRequestBody: true,
		SaveResponse:    false,
		MaxBodySize:     10240,
	})
}

// OperationLogWithConfig 带配置的操作日志中间件
func OperationLogWithConfig(config OperationLogConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 获取用户信息
		userID, _ := pkgContext.GetCurrentUserID(c)
		username := ""
		if user, err := pkgContext.GetCurrentUser(c); err == nil {
			username = user.Username
		}
		
		// 保存请求体
		var requestBody string
		if config.SaveRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 限制大小
				if config.MaxBodySize > 0 && len(bodyBytes) > config.MaxBodySize {
					requestBody = string(bodyBytes[:config.MaxBodySize]) + "...(truncated)"
				} else {
					requestBody = string(bodyBytes)
				}
				// 恢复 Body，供后续使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		
		// 保存响应体（如果需要）
		var responseBody string
		if config.SaveResponse {
			writer := &responseWriter{
				ResponseWriter: c.Writer,
				body:          &bytes.Buffer{},
			}
			c.Writer = writer
			defer func() {
				responseBody = writer.body.String()
				if config.MaxBodySize > 0 && len(responseBody) > config.MaxBodySize {
					responseBody = responseBody[:config.MaxBodySize] + "...(truncated)"
				}
			}()
		}
		
		// 执行请求
		c.Next()
		
		// 计算执行时长
		duration := time.Since(start).Milliseconds()
		
		// 提取错误信息
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
			if len(errorMsg) > 500 {
				errorMsg = errorMsg[:500]
			}
		}
		
		// 构建日志记录
		logEntry := OperationLogEntity{
			Module:       config.Module,
			Operation:    config.Operation,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			UserID:       userID,
			Username:     username,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  requestBody,
			ResponseBody: responseBody,
			Status:       c.Writer.Status(),
			Duration:     duration,
			ErrorMsg:     errorMsg,
		}
		
		// 异步保存日志（避免影响主流程性能）
		go saveOperationLog(logEntry)
	}
}

// saveOperationLog 保存操作日志到数据库
func saveOperationLog(log OperationLogEntity) {
	if err := database.DB.Create(&log).Error; err != nil {
		logger.Error("保存操作日志失败", zap.Error(err))
	}
}

// responseWriter 用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// OperationLogJSON 简化的 JSON 日志中间件（仅保存关键信息）
func OperationLogJSON(module, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		userID, _ := pkgContext.GetCurrentUserID(c)
		
		c.Next()
		
		duration := time.Since(start)
		
		// 使用结构化日志记录
		logger.Info(
			"[操作日志]",
			zap.String("module", module),
			zap.String("operation", operation),
			zap.Int64("userId", userID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// GetOperationLogList 获取操作日志列表（供业务层使用）
type OperationLogQuery struct {
	Module    string `form:"module"`
	Operation string `form:"operation"`
	Username  string `form:"username"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	PageNum   int    `form:"pageNum" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
}

// ToJSON 将日志转换为 JSON（用于响应体记录）
func (log *OperationLogEntity) ToJSON() string {
	data, _ := json.Marshal(log)
	return string(data)
}
