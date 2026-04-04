package logger

import (
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// getRequestID 从上下文获取 request-id
func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDHeader); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequestIDMiddleware 注入/生成 request-id
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Request.Header.Get(RequestIDHeader)
		if id == "" {
			id = generateUUID()
		}
		c.Set(RequestIDHeader, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// generateUUID 生成 UUID
func generateUUID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}

// Middleware access log 中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := getRequestID(c)
		l := Log.With(zap.String("requestId", reqID))
		c.Set("logger", l)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		l.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("clientIP", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}

// Recovery panic 日志
func Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(Log, true)
}