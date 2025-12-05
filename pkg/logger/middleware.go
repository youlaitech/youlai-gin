package logger

import (
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"youlai-gin/pkg/requestid"
)

// Middleware 统一 access log，自动注入 requestId
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := requestid.Get(c)
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

// Recovery 结构化 panic 日志
func Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(Log, true)
}

