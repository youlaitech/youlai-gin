package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/errs"
	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/redis"
)

const (
	defaultIPLimit     = 10 // 默认 IP 限流阈值（每秒请求数）
	rateLimitWindowSec = 1  // 限流窗口（秒）
)

// RateLimitByIP 基于 Redis 的 IP 限流中间件
// 对齐 youlai-boot RateLimiterFilter：Redis 固定窗口计数器，默认每秒 10 次/IP
func RateLimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := redis.RateLimiterIPPrefix + ip

		// Redis INCR 计数
		count, err := redis.Client.Incr(c.Request.Context(), key).Result()
		if err != nil {
			// Redis 异常时放行，避免影响正常请求
			c.Next()
			return
		}

		// 首次访问设置过期时间
		if count == 1 {
			redis.Client.Expire(c.Request.Context(), key, time.Duration(rateLimitWindowSec)*time.Second)
		}

		// 超过阈值则限流
		if count > defaultIPLimit {
			response.FromAppError(c, &errs.AppError{
				Code:       constant.CodeRequestConcurrencyLimitExceeded,
				Msg:        constant.MsgRequestConcurrencyLimitExceeded,
				HTTPStatus: http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}