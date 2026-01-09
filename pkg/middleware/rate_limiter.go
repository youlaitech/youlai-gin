package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"youlai-gin/pkg/constant"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/response"
)

// RateLimiter 限流器结构
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit // 每秒允许的请求数
	b        int        // 令牌桶容量
}

// NewRateLimiter 创建限流器
// r: 每秒允许的请求数
// b: 令牌桶容量（允许的突发请求数）
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// getLimiter 获取或创建指定 IP 的限流器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Middleware 限流中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			response.FromAppError(c, &errs.AppError{
				Code:       constant.CodeRequestConcurrencyLimitExceeded,
				Msg:        constant.MsgRequestConcurrencyLimitExceeded,
				HTTPStatus: http.StatusBadRequest,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupOldLimiters 定期清理不活跃的限流器（可选，防止内存泄漏）
func (rl *RateLimiter) CleanupOldLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			// 简单清理策略：清空所有限流器
			// 生产环境可以记录最后访问时间，只清理长时间未访问的
			rl.limiters = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}()
}

// RateLimitByIP 基于 IP 的限流中间件（快捷方式）
// 默认：每秒 10 个请求，突发 20 个
func RateLimitByIP() gin.HandlerFunc {
	limiter := NewRateLimiter(10, 20)
	// 每小时清理一次
	limiter.CleanupOldLimiters(time.Hour)
	return limiter.Middleware()
}

// RateLimitStrict 严格限流（用于敏感接口）
// 每秒 2 个请求，突发 5 个
func RateLimitStrict() gin.HandlerFunc {
	limiter := NewRateLimiter(2, 5)
	limiter.CleanupOldLimiters(time.Hour)
	return limiter.Middleware()
}
