package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const Header = "X-Request-ID"

// Middleware 注入/生成 request-id
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Request.Header.Get(Header)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(Header, id)
		c.Writer.Header().Set(Header, id)
		c.Next()
	}
}

// Get 读取 request-id
func Get(c *gin.Context) string {
	if v, ok := c.Get(Header); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

