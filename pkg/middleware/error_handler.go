package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/response"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		rawErr := c.Errors.Last().Err

		logger := extractLogger(c)

		// AppError
		if ae, ok := errs.As(rawErr); ok {
			// 打印底层错误（日志、trace）
			if ae.Err != nil {
				logger.Error("[ERROR]", zap.Error(ae.Err))
			}
			response.FromAppError(c, ae)
			return
		}

		// 未知错误 → 统一系统错误
		logger.Error("[SYSTEM ERROR]", zap.Error(rawErr))
		response.FromAppError(c, errs.SystemError(""))
	}
}

func extractLogger(c *gin.Context) *zap.Logger {
	if l, ok := c.Get("logger"); ok {
		if lg, ok := l.(*zap.Logger); ok {
			return lg
		}
	}
	return zap.L()
}
