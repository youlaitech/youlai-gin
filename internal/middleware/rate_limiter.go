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

		// Redis INCR 计数
		count, err := redis.Client.Incr(ctx, key).Result()
		if err != nil {
			// Redis 异常时放行，避免影响正常请求
			c.Next()
			return
		}

		// 确保一定有过期时间（防止 Expire 失败导致永久封禁）
		// 首次访问设置过期时间，或者发现 TTL 为 -1（无过期）时补设
		if count == 1 {
			if err := redis.Client.Expire(ctx, key, time.Duration(rateLimitWindowSec)*time.Second).Err(); err != nil {
				// Expire 失败时删除 key，避免永久封禁
				redis.Client.Del(ctx, key)
			}
		} else {
			// 兜底检查：如果 key 没有 TTL（TTL=-1），补设过期时间
			ttl, _ := redis.Client.TTL(ctx, key).Result()
			if ttl == -1 {
				redis.Client.Expire(ctx, key, time.Duration(rateLimitWindowSec)*time.Second)
			}
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
