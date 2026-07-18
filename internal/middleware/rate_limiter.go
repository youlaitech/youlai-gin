package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"youlai-gin/internal/common/config"
	"youlai-gin/internal/common/redis"
	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/errs"
	response "youlai-gin/internal/common"
)

const (
	defaultIPLimit     = 1000 // 默认 IP 限流阈值
	defaultIPWindowSec = 60   // 默认 IP 限流窗口（秒）
)

// SlidingWindowLua Redis ZSet 滑动窗口计数，返回窗口内累计请求数
const SlidingWindowLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local member = ARGV[3]
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
redis.call('ZADD', key, now, member)
redis.call('PEXPIRE', key, window + 1000)
return redis.call('ZCARD', key)
`

// RateLimitByIP IP 滑动窗口限流中间件
func RateLimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Cfg.RateLimit.Ip
		if !cfg.Enabled {
			c.Next()
			return
		}

		limit := cfg.Limit
		if limit <= 0 {
			limit = defaultIPLimit
		}

		window := cfg.Window
		if window == "" {
			window = "60s"
		}
		d, err := time.ParseDuration(window)
		if err != nil {
			d = defaultIPWindowSec * time.Second
		}
		windowMs := d.Milliseconds()
		windowSec := int64(windowMs / 1000)

		ip := c.ClientIP()
		key := redis.RateLimiterIPPrefix + ip
		ctx := c.Request.Context()

		now := time.Now().UnixMilli()
		member := uuid.New().String()

		count, err := redis.Client.Eval(ctx, SlidingWindowLua, []string{key}, now, windowMs, member).Int64()
		// Redis 异常时 Fail-Open：放行且不写限流响应头
		if err != nil {
			c.Next()
			return
		}

		// 窗口内剩余可用请求数
		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}
		// 限流窗口到期的 Unix 时间戳（秒）
		resetAt := time.Now().Unix() + windowSec

		if count > int64(limit) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))
			c.Header("Retry-After", strconv.FormatInt(windowSec, 10))
			response.FromAppError(c, &errs.AppError{
				Code:       constant.CodeRequestConcurrencyLimitExceeded,
				Msg:        constant.MsgRequestConcurrencyLimitExceeded,
				HTTPStatus: http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		// 正常放行：在 handler 之前写好限流响应头，随响应发出
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		c.Next()
	}
}
