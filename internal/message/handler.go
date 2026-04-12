package message

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/common/auth"
	response "youlai-gin/internal/common"
)

type SseHandler struct {
	sseService  *SseService
	tokenParser auth.TokenManager
}

func NewSseHandler(tokenParser auth.TokenManager) *SseHandler {
	return &SseHandler{
		sseService:  GetSseService(),
		tokenParser: tokenParser,
	}
}

// Connect SSE连接接口
// @Summary 建立SSE连接
// @Tags 13.SSE连接
// @Accept json
// @Produce text/event-stream
// @Param token query string false "JWT Token (可选，优先使用Authorization Header)"
// @Success 200 {object} string "SSE Stream"
// @Router /api/v1/sse/connect [get]
func (h *SseHandler) Connect(c *gin.Context) {
	// 从 Authorization 头或 query 参数获取 token
	tokenString := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		tokenString = c.Query("token")
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "未授权"})
		return
	}

	// 解析 token 获取用户信息
	userDetails, err := h.tokenParser.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "无效的Token"})
		return
	}

	username := userDetails.Username

	if username == "" {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "无效的用户信息"})
		return
	}

	// 创建 SSE 连接
	emitter, err := h.sseService.CreateConnection(username, c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Result{Code: "B0001", Msg: "SSE连接创建失败"})
		return
	}

	// 启动心跳协程
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.Request.Context().Done():
				return
			case <-ticker.C:
				if err := emitter.SendHeartbeat(); err != nil {
					return
				}
			}
		}
	}()

	// 同时监听客户端断开和服务关闭信号
	select {
	case <-c.Request.Context().Done():
		// 客户端主动断开连接
	case <-emitter.Done():
		// 服务端主动关闭（如应用停止）
	}
	h.sseService.RemoveEmitter(emitter)
}

// GetOnlineCount 获取在线用户数
// @Summary 获取在线用户数
// @Tags 13.SSE连接
// @Accept json
// @Produce json
// @Success 200 {object} response.Result
// @Router /api/v1/sse/online-count [get]
func (h *SseHandler) GetOnlineCount(c *gin.Context) {
	c.JSON(http.StatusOK, response.Result{Code: "00000", Msg: "操作成功", Data: h.sseService.GetOnlineUserCount()})
}

func RegisterRoutes(r *gin.RouterGroup, tokenParser auth.TokenManager) {
	handler := NewSseHandler(tokenParser)
	sseGroup := r.Group("/sse")
	{
		sseGroup.GET("/connect", handler.Connect)
		sseGroup.GET("/online-count", handler.GetOnlineCount)
	}
}
