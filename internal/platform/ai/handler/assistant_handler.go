package handler

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/ai/model"
	"youlai-gin/internal/platform/ai/service"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// ParseCommand 解析自然语言命令
func ParseCommand(c *gin.Context) {
	var req model.AiParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	user, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	ip := clientIP(c)
	result, err := service.ParseCommand(c.Request.Context(), &req, user.UserID, user.Username, ip)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// ExecuteCommand 执行已解析的命令
func ExecuteCommand(c *gin.Context) {
	var req model.AiExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	user, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	ip := clientIP(c)
	result, err := service.ExecuteCommand(&req, user.UserID, user.Username, ip)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// clientIP 获取客户端 IP
func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip != "" {
		return ip
	}

	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			candidate := strings.TrimSpace(parts[0])
			if candidate != "" {
				return candidate
			}
		}
	}

	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}
