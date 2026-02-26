package stomp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StompHandler STOMP WebSocket 处理器
type StompHandler struct {
	broker *StompBroker
	// authFunc 用户认证函数，从请求中获取用户信息
	authFunc func(c *gin.Context) (userId int64, username string, err error)
}

// NewStompHandler 创建 STOMP 处理器
func NewStompHandler(broker *StompBroker, authFunc func(c *gin.Context) (int64, string, error)) *StompHandler {
	return &StompHandler{
		broker:   broker,
		authFunc: authFunc,
	}
}

// HandleWebSocket 处理 WebSocket 连接请求
// @Summary STOMP WebSocket 连接
// @Description 建立 STOMP over WebSocket 连接
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 101 "Switching Protocols"
// @Router /ws [get]
func (h *StompHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户信息
	userId, username, err := h.authFunc(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized: " + err.Error(),
		})
		return
	}

	// 处理 WebSocket 升级和 STOMP 连接
	if err := h.broker.ServeHTTP(c.Writer, c.Request, userId, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "WebSocket upgrade failed: " + err.Error(),
		})
		return
	}
}

// GetOnlineCount 获取在线用户数
// @Summary 获取在线用户数
// @Description 获取当前在线用户数量
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ws/online-count [get]
func (h *StompHandler) GetOnlineCount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    h.broker.GetOnlineUserCount(),
		"message": "success",
	})
}
