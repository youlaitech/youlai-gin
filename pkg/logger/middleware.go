package logger

import (
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"youlai-gin/pkg/requestid"
)

// Middleware access log 中间件
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

// Recovery panic 日志
func Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(Log, true)
}

