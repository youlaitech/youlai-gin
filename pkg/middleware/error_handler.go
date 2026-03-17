package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/response"
)

// ErrorHandler 统一错误处理中间件，类似 Java Spring 的 @ControllerAdvice + @ExceptionHandler
// 处理 c.Error(err) 注册的错误，配合 Recovery() 处理 panic
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		rawErr := c.Errors.Last().Err
		logger := extractLogger(c)

		// AppError 业务错误
		if ae, ok := errs.As(rawErr); ok {
			// 记录底层错误日志
			if ae.Err != nil {
				logger.Error("[BUSINESS ERROR]",
					zap.String("code", ae.Code),
					zap.String("msg", ae.Msg),
					zap.Error(ae.Err),
					zap.String("path", c.Request.URL.Path),
				)
			}
			// 设置 HTTP 状态码并返回响应
			c.Status(ae.HTTPStatus)
			response.FromAppError(c, ae)
			return
		}

		// 未知错误统一返回系统错误
		logger.Error("[UNKNOWN ERROR]",
			zap.Error(rawErr),
			zap.String("path", c.Request.URL.Path),
		)
		c.Status(http.StatusBadRequest)
		response.FromAppError(c, errs.SystemError(""))
	}
}

// extractLogger 从上下文提取 logger，降级使用全局 logger
func extractLogger(c *gin.Context) *zap.Logger {
	if l, ok := c.Get("logger"); ok {
		if lg, ok := l.(*zap.Logger); ok {
			return lg
		}
	}
	return zap.L()
}
