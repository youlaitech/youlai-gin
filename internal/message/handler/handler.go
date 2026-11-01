package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/logger"
	msgModel "youlai-gin/internal/message/model"
	realtimeService "youlai-gin/internal/message/service"
)

// 心跳间隔与超时，与 youlai-boot 保持一致
const (
	pingInterval  = 25 * time.Second
	pongTimeout   = 60 * time.Second
	maxFrameBytes = 16 * 1024
)

// RealtimeHandler 实时连接接口层
type RealtimeHandler struct {
	realtimeService *realtimeService.RealtimeService
	tokenParser     auth.TokenManager
	upgrader        websocket.Upgrader
}

// createTicketRequest 申请握手票据的请求体
type createTicketRequest struct {
	ClientInstanceID string   `json:"clientInstanceId"`
	Platform         string   `json:"platform"`
	Topics           []string `json:"topics"`
}

// NewRealtimeHandler 创建 RealtimeHandler 实例
func NewRealtimeHandler(tokenParser auth.TokenManager) *RealtimeHandler {
	return &RealtimeHandler{
		realtimeService: realtimeService.GetRealtimeService(),
		tokenParser:     tokenParser,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     sameOrigin,
		},
	}
}

// RegisterRoutes 注册实时连接路由，基础路径 /api/v1/realtime
func (h *RealtimeHandler) RegisterRoutes(r *gin.RouterGroup) {
	realtimeGroup := r.Group("/realtime")
	realtimeGroup.POST("/tickets", h.CreateTicket)
	realtimeGroup.GET("/presence", h.GetPresence)
	realtimeGroup.GET("/ws", h.Connect)
}

// CreateTicket 签发一次性握手票据
// @Summary 签发实时连接票据
// @Tags 13.实时推送
// @Accept json
// @Produce json
// @Success 200 {object} response.Result
// @Router /api/v1/realtime/tickets [post]
func (h *RealtimeHandler) CreateTicket(c *gin.Context) {
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Result{Code: "A0400", Msg: "参数校验失败"})
		return
	}
	if !containsTopic(req.Topics, msgModel.TopicNotice) {
		c.JSON(http.StatusBadRequest, response.Result{Code: "A0400", Msg: "订阅主题须包含 notice"})
		return
	}
	for _, topic := range req.Topics {
		if _, ok := msgModel.SupportedTopics[topic]; !ok {
			c.JSON(http.StatusBadRequest, response.Result{Code: "A0400", Msg: "不支持的订阅主题：" + topic})
			return
		}
	}

	tokenString := bearerToken(c)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "未授权"})
		return
	}
	userDetails, err := h.tokenParser.ParseToken(tokenString)
	if err != nil || userDetails.Username == "" {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "无效的Token"})
		return
	}

	ticket, err := h.realtimeService.CreateTicket(userDetails.Username, tokenString, req.Topics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Result{Code: "B0001", Msg: "票据签发失败"})
		return
	}
	c.JSON(http.StatusOK, response.Result{Code: "00000", Msg: "操作成功", Data: gin.H{"ticket": ticket}})
}

// GetPresence 查询在线状态
// @Summary 查询在线状态
// @Tags 13.实时推送
// @Produce json
// @Success 200 {object} response.Result
// @Router /api/v1/realtime/presence [get]
func (h *RealtimeHandler) GetPresence(c *gin.Context) {
	c.JSON(http.StatusOK, response.Result{
		Code: "00000",
		Msg:  "操作成功",
		Data: h.realtimeService.PresenceSnapshot(),
	})
}

// Connect 建立实时连接，握手只接受一次性票据
// @Summary 建立实时连接
// @Tags 13.实时推送
// @Param ticket query string true "一次性握手票据"
// @Success 101 {string} string "Switching Protocols"
func (h *RealtimeHandler) Connect(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "缺少握手票据"})
		return
	}
	grant, err := h.realtimeService.ConsumeTicket(ticket)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "票据无效或已使用"})
		return
	}
	if _, err := h.tokenParser.ParseToken(grant.AccessToken); err != nil {
		c.JSON(http.StatusUnauthorized, response.Result{Code: "A0401", Msg: "访问凭证已失效"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warn("WebSocket 升级失败", zap.Error(err))
		return
	}

	topics := make(map[string]struct{}, len(grant.Topics))
	for _, topic := range grant.Topics {
		topics[topic] = struct{}{}
	}
	client := realtimeService.NewRealtimeClient(conn, grant.Username, topics)
	h.realtimeService.RegisterClient(client)
	defer h.realtimeService.RemoveClient(client)

	go h.writePump(client, conn)
	go h.heartbeatPump(client, c.Request.Context().Done())

	_ = client.Send(msgModel.NewRealtimeMessage(msgModel.EventConnectionReady, gin.H{"heartbeatTimeoutMs": msgModel.HeartbeatTimeoutMs}))
	h.realtimeService.SendOnlineCount()

	h.readPump(client, conn)
}

// writePump 把发送队列中的消息写入连接
func (h *RealtimeHandler) writePump(client *realtimeService.RealtimeClient, conn *websocket.Conn) {
	for {
		select {
		case <-client.Done():
			return
		case data := <-client.SendChan():
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				client.Close()
				return
			}
		}
	}
}

// heartbeatPump 定时发送心跳，超时未响应则断开连接
func (h *RealtimeHandler) heartbeatPump(client *realtimeService.RealtimeClient, requestDone <-chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-client.Done():
			return
		case <-requestDone:
			return
		case <-ticker.C:
			if time.Since(client.LastPongAt()) >= pongTimeout {
				client.Close()
				return
			}
			ping := msgModel.NewRealtimeMessage(msgModel.EventHeartbeatPing, map[string]interface{}{})
			client.ExpectPong(ping.ID)
			if err := client.Send(ping); err != nil {
				client.Close()
				return
			}
		}
	}
}

// readPump 读取客户端消息，只处理心跳响应
func (h *RealtimeHandler) readPump(client *realtimeService.RealtimeClient, conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if len(data) > maxFrameBytes {
			client.Close()
			break
		}
		if pingID, ok := msgModel.ParsePong(data); ok {
			client.ConfirmPong(pingID)
		}
	}
}

// bearerToken 从 Authorization 头读取访问令牌
func bearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return authHeader[7:]
	}
	return ""
}

// containsTopic 判断主题是否在订阅列表中
func containsTopic(topics []string, target string) bool {
	for _, topic := range topics {
		if topic == target {
			return true
		}
	}
	return false
}

// sameOrigin 校验握手来源，浏览器连接必须同源
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return strings.EqualFold(origin, "http://"+r.Host) || strings.EqualFold(origin, "https://"+r.Host)
}
