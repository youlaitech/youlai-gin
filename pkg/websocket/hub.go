package websocket

import (
	"encoding/json"
	"sync"
	
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	
	"youlai-gin/pkg/logger"
)

// Hub WebSocket 连接管理中心
type Hub struct {
	// 所有客户端连接（key: userID）
	clients map[int64]*Client
	
	// 广播消息通道
	broadcast chan *Message
	
	// 注册客户端通道
	register chan *Client
	
	// 注销客户端通道
	unregister chan *Client
	
	// 互斥锁
	mu sync.RWMutex
}

// Client WebSocket 客户端
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID int64
	send   chan *Message
}

// Message 消息结构
type Message struct {
	Type    string      `json:"type"`    // 消息类型：notice, message, system
	Title   string      `json:"title"`   // 标题
	Content string      `json:"content"` // 内容
	Data    interface{} `json:"data"`    // 附加数据
	UserIDs []int64     `json:"userIds"` // 目标用户ID列表（为空则广播）
}

// DefaultHub 默认Hub实例
var DefaultHub *Hub

// InitHub 初始化Hub
func InitHub() {
	DefaultHub = &Hub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go DefaultHub.Run()
}

// Run 启动Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			logger.Info("WebSocket 客户端已连接", zap.Int64("userID", client.userID), zap.Int("online", len(h.clients)))
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				logger.Info("WebSocket 客户端已断开", zap.Int64("userID", client.userID), zap.Int("online", len(h.clients)))
			}
			h.mu.Unlock()
			
		case message := <-h.broadcast:
			h.mu.RLock()
			if len(message.UserIDs) == 0 {
				// 广播给所有客户端
				for _, client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client.userID)
					}
				}
			} else {
				// 发送给指定用户
				for _, userID := range message.UserIDs {
					if client, ok := h.clients[userID]; ok {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, userID)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendMessage 发送消息给指定用户
func (h *Hub) SendMessage(userIDs []int64, message *Message) {
	message.UserIDs = userIDs
	h.broadcast <- message
}

// BroadcastMessage 广播消息给所有用户
func (h *Hub) BroadcastMessage(message *Message) {
	message.UserIDs = nil
	h.broadcast <- message
}

// GetOnlineCount 获取在线人数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsOnline 检查用户是否在线
func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// readPump 读取客户端消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket 读取错误", zap.Error(err))
			}
			break
		}
		
		// 处理客户端消息（如心跳）
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
				// 响应心跳
				c.send <- &Message{Type: "pong"}
			}
		}
	}
}

// writePump 向客户端写入消息
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		
		data, err := json.Marshal(message)
		if err != nil {
			logger.Error("WebSocket 消息序列化失败", zap.Error(err))
			continue
		}
		
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Error("WebSocket 写入错误", zap.Error(err))
			return
		}
	}
}

// ServeWs 处理WebSocket连接
func ServeWs(hub *Hub, conn *websocket.Conn, userID int64) {
	client := &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan *Message, 256),
	}
	
	client.hub.register <- client
	
	// 启动读写协程
	go client.writePump()
	go client.readPump()
}
