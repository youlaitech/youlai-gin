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
func RateLimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := redis.RateLimiterIPPrefix + ip
		ctx := c.Request.Context()

		count, err := redis.Client.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			if err := redis.Client.Expire(ctx, key, time.Duration(rateLimitWindowSec)*time.Second).Err(); err != nil {
				redis.Client.Del(ctx, key)
			}
		} else {
			ttl, _ := redis.Client.TTL(ctx, key).Result()
			if ttl == -1 {
				redis.Client.Expire(ctx, key, time.Duration(rateLimitWindowSec)*time.Second)
			}
		}

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
