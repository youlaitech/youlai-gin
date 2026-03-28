package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"youlai-gin/pkg/database"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/logger"
	"youlai-gin/pkg/enums"
)

// OperationLogEntity 操作日志实体
type OperationLogEntity struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Module        int        `gorm:"column:module" json:"module"`
	ActionType    int        `gorm:"column:action_type" json:"actionType"`
	Title         string     `gorm:"column:title;size:100" json:"title"`
	Content       string     `gorm:"column:content;type:text" json:"content"`
	OperatorID    int64      `gorm:"column:operator_id" json:"operatorId"`
	OperatorName  string     `gorm:"column:operator_name;size:50" json:"operatorName"`
	RequestURI    string     `gorm:"column:request_uri;size:255" json:"requestUri"`
	RequestMethod string     `gorm:"column:request_method;size:10" json:"requestMethod"`
	IP            string     `gorm:"column:ip;size:45" json:"ip"`
	Province      string     `gorm:"column:province;size:100" json:"province"`
	City          string     `gorm:"column:city;size:100" json:"city"`
	Device        string     `gorm:"column:device;size:100" json:"device"`
	OS            string     `gorm:"column:os;size:100" json:"os"`
	Browser       string     `gorm:"column:browser;size:100" json:"browser"`
	Status        int        `gorm:"column:status" json:"status"`
	ErrorMsg      string     `gorm:"column:error_msg;size:255" json:"errorMsg"`
	ExecutionTime int        `gorm:"column:execution_time" json:"executionTime"`
	CreateBy      int64      `gorm:"column:create_by" json:"createBy"`
	CreateTime    time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (OperationLogEntity) TableName() string {
	return "sys_log"
}

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Module           enums.LogModule
	ActionType       enums.ActionType
	Title            string
	Content          string
	SaveRequestBody  bool
	SaveResponse     bool
	MaxBodySize      int
}

// DefaultOperationLogConfig 默认配置
var DefaultOperationLogConfig = OperationLogConfig{
	SaveRequestBody: true,
	SaveResponse:    false,
	MaxBodySize:     10240,
}

// OperationLog 操作日志中间件
func OperationLog(module enums.LogModule, actionType enums.ActionType) gin.HandlerFunc {
	return OperationLogWithConfig(OperationLogConfig{
		Module:          module,
		ActionType:      actionType,
		SaveRequestBody: true,
		SaveResponse:    false,
		MaxBodySize:     10240,
	})
}

// OperationLogWithConfig 带配置的操作日志中间件
func OperationLogWithConfig(config OperationLogConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		userID, _ := pkgContext.GetCurrentUserID(c)
		var username string
		if user, err := pkgContext.GetCurrentUser(c); err == nil {
			username = user.Username
		}

		if config.SaveRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

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

		c.Next()

		ua := c.Request.UserAgent()
		duration := time.Since(start).Milliseconds()

		var errorMsg string
		status := 1
		if len(c.Errors) > 0 {
			status = 0
			errorMsg = c.Errors.String()
			if len(errorMsg) > 255 {
				errorMsg = errorMsg[:255]
			}
		}

		module := config.Module
		if module == 0 {
			module = enums.LogModuleOther
		}

		actionType := config.ActionType
		if actionType == 0 {
			actionType = enums.ActionTypeOther
		}

		title := config.Title
		if title == "" {
			title = enums.LogModuleDesc[module] + "-" + enums.ActionTypeDesc[actionType]
		}

		logEntry := OperationLogEntity{
			Module:        int(module),
			ActionType:    int(actionType),
			Title:         title,
			Content:       config.Content,
			OperatorID:    userID,
			OperatorName:  username,
			RequestURI:    c.Request.URL.Path,
			RequestMethod: c.Request.Method,
			IP:            c.ClientIP(),
			Province:      "",
			City:          "",
			Device:        "",
			OS:            ParseOS(ua),
			Browser:       ParseBrowser(ua),
			Status:        status,
			ErrorMsg:      errorMsg,
			ExecutionTime: int(duration),
		}

		go saveOperationLog(logEntry)
	}
}

// saveOperationLog 保存操作日志到数据库
func saveOperationLog(log OperationLogEntity) {
	if err := database.DB.Create(&log).Error; err != nil {
		logger.Error("保存操作日志失败", zap.Error(err))
	}
}

// SaveOperationLog 导出的保存操作日志函数
func SaveOperationLog(log OperationLogEntity) error {
	return database.DB.Create(&log).Error
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

// OperationLogJSON JSON 日志中间件
func OperationLogJSON(actionType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		userID, _ := pkgContext.GetCurrentUserID(c)

		c.Next()

		duration := time.Since(start)

		logger.Info(
			"[操作日志]",
			zap.String("actionType", actionType),
			zap.Int64("userId", userID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// OperationLogQuery 操作日志查询参数
type OperationLogQuery struct {
	ActionType string `form:"actionType"`
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	PageNum    int    `form:"pageNum" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
}

// ParseBrowser 从 User-Agent 字符串中提取浏览器名称和版本
func ParseBrowser(ua string) string {
	re := regexp.MustCompile(`(?:Edg|OPR|Opera|Firefox|Chrome|Safari|Version|MSIE|Trident)/[\d.]+`)
	matches := re.FindAllString(ua, -1)
	browser := ""

	switch {
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera/"):
		browser = "Opera"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Chrome/") && !strings.Contains(ua, "Edg/"):
		browser = "Chrome"
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome"):
		browser = "Safari"
	case strings.Contains(ua, "MSIE") || strings.Contains(ua, "Trident/"):
		browser = "IE"
	}

	// 提取版本号
	for _, m := range matches {
		if strings.HasPrefix(m, browser) && strings.Contains(m, "/") {
			if idx := strings.Index(m, "/"); idx != -1 {
				browser = browser + " " + m[idx+1:]
			}
			break
		}
		// Safari 的版本号在 Version/ 后面
		if browser == "Safari" && strings.HasPrefix(m, "Version/") {
			if idx := strings.Index(m, "/"); idx != -1 {
				browser = "Safari " + m[idx+1:]
			}
			break
		}
	}

	return browser
}

// ParseOS 从 User-Agent 字符串中提取操作系统
func ParseOS(ua string) string {
	switch {
	case strings.Contains(ua, "Windows NT 10"):
		return "Windows 10"
	case strings.Contains(ua, "Windows NT 6.3"):
		return "Windows 8.1"
	case strings.Contains(ua, "Windows NT 6.1"):
		return "Windows 7"
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "Mac OS X"):
		re := regexp.MustCompile(`Mac OS X ([\d_]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "macOS " + strings.ReplaceAll(matches[1], "_", ".")
		}
		return "macOS"
	case strings.Contains(ua, "Android"):
		re := regexp.MustCompile(`Android ([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Android " + matches[1]
		}
		return "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		re := regexp.MustCompile(`OS ([\d_]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "iOS " + strings.ReplaceAll(matches[1], "_", ".")
		}
		return "iOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return ""
	}
}
