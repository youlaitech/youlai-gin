package stomp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"youlai-gin/pkg/logger"
)

// StompBroker STOMP 消息代理
type StompBroker struct {
	// 连接管理
	connections   map[string]*StompConnection // connectionId -> connection
	userConns     map[int64]map[string]*StompConnection // userId -> connections
	subscriptions map[string]map[string]*StompConnection // destination -> connections

	// 计数器
	messageCounter int64

	// 互斥锁
	mu sync.RWMutex

	// 配置
	config *BrokerConfig
}

// BrokerConfig 代理配置
type BrokerConfig struct {
	// 心跳配置
	SendHeartBeatInterval time.Duration // 发送心跳间隔
	RecvHeartBeatTimeout  time.Duration // 接收心跳超时

	// 连接配置
	MaxConnections    int           // 最大连接数
	ReadBufferSize    int           // 读缓冲区大小
	WriteBufferSize   int           // 写缓冲区大小
	EnableCompression bool          // 启用压缩
	WriteTimeout      time.Duration // 写超时
	ReadTimeout       time.Duration // 读超时
}

// DefaultBrokerConfig 默认配置
func DefaultBrokerConfig() *BrokerConfig {
	return &BrokerConfig{
		SendHeartBeatInterval: 10 * time.Second,
		RecvHeartBeatTimeout:  30 * time.Second,
		MaxConnections:        10000,
		ReadBufferSize:        1024,
		WriteBufferSize:       1024,
		EnableCompression:     false,
		WriteTimeout:          10 * time.Second,
		ReadTimeout:           30 * time.Second,
	}
}

// StompConnection STOMP 连接
type StompConnection struct {
	// 连接信息
	connectionId string
	userId       int64
	username     string

	// WebSocket 连接
	wsConn *websocket.Conn

	// 订阅管理
	subscriptions map[string]string // subscriptionId -> destination

	// 所属代理
	broker *StompBroker

	// 上下文控制
	ctx    context.Context
	cancel context.CancelFunc

	// 发送消息通道
	sendChan chan []byte

	// 最后活动时间
	lastActivity time.Time

	// 心跳配置
	sendHeartBeatInterval time.Duration
	recvHeartBeatTimeout  time.Duration
}

// NewStompBroker 创建 STOMP 代理
func NewStompBroker(config *BrokerConfig) *StompBroker {
	if config == nil {
		config = DefaultBrokerConfig()
	}

	return &StompBroker{
		connections:   make(map[string]*StompConnection),
		userConns:     make(map[int64]map[string]*StompConnection),
		subscriptions: make(map[string]map[string]*StompConnection),
		config:        config,
	}
}

// ServeHTTP 处理 WebSocket 升级和 STOMP 连接
func (b *StompBroker) ServeHTTP(w http.ResponseWriter, r *http.Request, userId int64, username string) error {
	// WebSocket 升级器
	upgrader := websocket.Upgrader{
		ReadBufferSize:    b.config.ReadBufferSize,
		WriteBufferSize:   b.config.WriteBufferSize,
		EnableCompression: b.config.EnableCompression,
		CheckOrigin: func(r *http.Request) bool {
			return true // 生产环境应根据需要配置
		},
	}

	// 升级 HTTP 连接到 WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	// 创建 STOMP 连接
	conn := b.createConnection(wsConn, userId, username)

	// 启动连接处理
	go conn.run()

	return nil
}

// createConnection 创建 STOMP 连接
func (b *StompBroker) createConnection(wsConn *websocket.Conn, userId int64, username string) *StompConnection {
	ctx, cancel := context.WithCancel(context.Background())

	conn := &StompConnection{
		connectionId:          generateConnectionId(),
		userId:                userId,
		username:              username,
		wsConn:                wsConn,
		subscriptions:         make(map[string]string),
		broker:                b,
		ctx:                   ctx,
		cancel:                cancel,
		sendChan:              make(chan []byte, 256),
		lastActivity:          time.Now(),
		sendHeartBeatInterval: b.config.SendHeartBeatInterval,
		recvHeartBeatTimeout:  b.config.RecvHeartBeatTimeout,
	}

	// 注册连接
	b.registerConnection(conn)

	return conn
}

// registerConnection 注册连接
func (b *StompBroker) registerConnection(conn *StompConnection) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.connections[conn.connectionId] = conn

	if conn.userId > 0 {
		if b.userConns[conn.userId] == nil {
			b.userConns[conn.userId] = make(map[string]*StompConnection)
		}
		b.userConns[conn.userId][conn.connectionId] = conn
	}

	logger.Info("STOMP 连接已建立",
		zap.String("connectionId", conn.connectionId),
		zap.Int64("userId", conn.userId),
		zap.String("username", conn.username),
		zap.Int("totalConnections", len(b.connections)),
	)
}

// unregisterConnection 注销连接
func (b *StompBroker) unregisterConnection(conn *StompConnection) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 移除所有订阅
	for subId, destination := range conn.subscriptions {
		if conns, ok := b.subscriptions[destination]; ok {
			delete(conns, conn.connectionId)
			if len(conns) == 0 {
				delete(b.subscriptions, destination)
			}
		}
		_ = subId
	}

	// 移除连接
	delete(b.connections, conn.connectionId)

	if conn.userId > 0 {
		if conns, ok := b.userConns[conn.userId]; ok {
			delete(conns, conn.connectionId)
			if len(conns) == 0 {
				delete(b.userConns, conn.userId)
			}
		}
	}

	logger.Info("STOMP 连接已断开",
		zap.String("connectionId", conn.connectionId),
		zap.Int64("userId", conn.userId),
		zap.Int("totalConnections", len(b.connections)),
	)
}

// Subscribe 订阅主题
func (b *StompBroker) Subscribe(conn *StompConnection, subscriptionId, destination string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 记录连接的订阅
	conn.subscriptions[subscriptionId] = destination

	// 添加到主题订阅列表
	if b.subscriptions[destination] == nil {
		b.subscriptions[destination] = make(map[string]*StompConnection)
	}
	b.subscriptions[destination][conn.connectionId] = conn

	logger.Debug("STOMP 订阅成功",
		zap.String("connectionId", conn.connectionId),
		zap.String("subscriptionId", subscriptionId),
		zap.String("destination", destination),
	)
}

// Unsubscribe 取消订阅
func (b *StompBroker) Unsubscribe(conn *StompConnection, subscriptionId string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	destination, ok := conn.subscriptions[subscriptionId]
	if !ok {
		return
	}

	// 从连接的订阅中移除
	delete(conn.subscriptions, subscriptionId)

	// 从主题订阅列表中移除
	if conns, ok := b.subscriptions[destination]; ok {
		delete(conns, conn.connectionId)
		if len(conns) == 0 {
			delete(b.subscriptions, destination)
		}
	}

	logger.Debug("STOMP 取消订阅",
		zap.String("connectionId", conn.connectionId),
		zap.String("subscriptionId", subscriptionId),
		zap.String("destination", destination),
	)
}

// Broadcast 广播消息到主题
func (b *StompBroker) Broadcast(destination string, payload interface{}) error {
	b.mu.RLock()
	conns, ok := b.subscriptions[destination]
	if !ok || len(conns) == 0 {
		b.mu.RUnlock()
		return nil
	}

	// 复制连接列表避免长时间持锁
	connList := make([]*StompConnection, 0, len(conns))
	for _, conn := range conns {
		connList = append(connList, conn)
	}
	b.mu.RUnlock()

	// 序列化消息体
	var body []byte
	switch v := payload.(type) {
	case string:
		body = []byte(v)
	case []byte:
		body = v
	default:
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	// 发送消息给所有订阅者
	for _, conn := range connList {
		if subscriptionId, ok := conn.subscriptions[destination]; ok {
			b.mu.RLock()
			b.messageCounter++
			messageId := fmt.Sprintf("msg-%d", b.messageCounter)
			b.mu.RUnlock()

			frame := NewMessageFrame(destination, subscriptionId, messageId, body)
			conn.SendFrame(frame)
		}
	}

	return nil
}

// SendToUser 发送消息给指定用户
func (b *StompBroker) SendToUser(userId int64, destination string, payload interface{}) error {
	b.mu.RLock()
	conns, ok := b.userConns[userId]
	if !ok || len(conns) == 0 {
		b.mu.RUnlock()
		return nil
	}

	// 复制连接列表
	connList := make([]*StompConnection, 0, len(conns))
	for _, conn := range conns {
		connList = append(connList, conn)
	}
	b.mu.RUnlock()

	// 序列化消息体
	var body []byte
	switch v := payload.(type) {
	case string:
		body = []byte(v)
	case []byte:
		body = v
	default:
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	// 发送消息给用户的所有连接
	for _, conn := range connList {
		if subscriptionId, ok := conn.subscriptions[destination]; ok {
			b.mu.RLock()
			b.messageCounter++
			messageId := fmt.Sprintf("msg-%d", b.messageCounter)
			b.mu.RUnlock()

			frame := NewMessageFrame(destination, subscriptionId, messageId, body)
			conn.SendFrame(frame)
		}
	}

	return nil
}

// GetOnlineUserCount 获取在线用户数
func (b *StompBroker) GetOnlineUserCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.userConns)
}

// GetTotalConnectionCount 获取总连接数
func (b *StompBroker) GetTotalConnectionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.connections)
}

// run 运行连接处理循环
func (c *StompConnection) run() {
	defer func() {
		c.Close()
		c.broker.unregisterConnection(c)
	}()

	// 启动写协程
	go c.writePump()

	// 启动心跳协程
	go c.heartBeatPump()

	// 读取消息循环
	c.readPump()
}

// readPump 读取消息循环
func (c *StompConnection) readPump() {
	defer c.cancel()

	// 设置读超时
	c.wsConn.SetReadLimit(65536)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 设置读超时
			c.wsConn.SetReadDeadline(time.Now().Add(c.recvHeartBeatTimeout))

			// 读取 WebSocket 消息
			messageType, data, err := c.wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket 读取错误", zap.Error(err))
				}
				return
			}

			// 只处理文本消息
			if messageType != websocket.TextMessage {
				continue
			}

			// 更新最后活动时间
			c.lastActivity = time.Now()

			// 处理心跳（单个换行符）
			if len(data) == 1 && data[0] == '\n' {
				continue
			}

			// 解析 STOMP 帧
			frame, err := Unmarshal(data)
			if err != nil {
				logger.Error("STOMP 帧解析失败", zap.Error(err))
				c.SendError("Frame parse error: " + err.Error())
				continue
			}

			// 处理 STOMP 命令
			c.handleFrame(frame)
		}
	}
}

// writePump 写消息循环
func (c *StompConnection) writePump() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.sendChan:
			c.wsConn.SetWriteDeadline(time.Now().Add(c.broker.config.WriteTimeout))
			if err := c.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
				logger.Error("WebSocket 写入错误", zap.Error(err))
				return
			}
		}
	}
}

// heartBeatPump 心跳发送
func (c *StompConnection) heartBeatPump() {
	ticker := time.NewTicker(c.sendHeartBeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// 检查是否超时
			if time.Since(c.lastActivity) > c.recvHeartBeatTimeout {
				logger.Warn("STOMP 心跳超时，关闭连接",
					zap.String("connectionId", c.connectionId),
				)
				c.cancel()
				return
			}

			// 发送心跳
			c.sendChan <- []byte("\n")
		}
	}
}

// handleFrame 处理 STOMP 帧
func (c *StompConnection) handleFrame(frame *Frame) {
	switch frame.Command {
	case CmdConnect:
		c.handleConnect(frame)
	case CmdSubscribe:
		c.handleSubscribe(frame)
	case CmdUnsubscribe:
		c.handleUnsubscribe(frame)
	case CmdSend:
		c.handleSend(frame)
	case CmdDisconnect:
		c.handleDisconnect(frame)
	case CmdAck, CmdNack:
		// 暂不实现 ACK/NACK
	default:
		logger.Warn("未知的 STOMP 命令",
			zap.String("command", frame.Command),
			zap.String("connectionId", c.connectionId),
		)
	}
}

// handleConnect 处理 CONNECT 命令
func (c *StompConnection) handleConnect(frame *Frame) {
	// 发送 CONNECTED 帧
	connectedFrame := NewConnectedFrame()
	c.SendFrame(connectedFrame)

	logger.Debug("STOMP CONNECT 处理完成",
		zap.String("connectionId", c.connectionId),
	)
}

// handleSubscribe 处理 SUBSCRIBE 命令
func (c *StompConnection) handleSubscribe(frame *Frame) {
	destination := frame.GetHeader(HdrDestination)
	subscriptionId := frame.GetHeader(HdrId)

	if destination == "" || subscriptionId == "" {
		c.SendError("SUBSCRIBE requires 'destination' and 'id' headers")
		return
	}

	c.broker.Subscribe(c, subscriptionId, destination)

	// 发送收据（如果请求）
	if receipt := frame.GetHeader(HdrReceipt); receipt != "" {
		c.SendReceipt(receipt)
	}
}

// handleUnsubscribe 处理 UNSUBSCRIBE 命令
func (c *StompConnection) handleUnsubscribe(frame *Frame) {
	subscriptionId := frame.GetHeader(HdrId)

	if subscriptionId == "" {
		c.SendError("UNSUBSCRIBE requires 'id' header")
		return
	}

	c.broker.Unsubscribe(c, subscriptionId)

	// 发送收据（如果请求）
	if receipt := frame.GetHeader(HdrReceipt); receipt != "" {
		c.SendReceipt(receipt)
	}
}

// handleSend 处理 SEND 命令
func (c *StompConnection) handleSend(frame *Frame) {
	destination := frame.GetHeader(HdrDestination)

	if destination == "" {
		c.SendError("SEND requires 'destination' header")
		return
	}

	// 广播消息到目标主题
	c.broker.Broadcast(destination, frame.Body)

	// 发送收据（如果请求）
	if receipt := frame.GetHeader(HdrReceipt); receipt != "" {
		c.SendReceipt(receipt)
	}
}

// handleDisconnect 处理 DISCONNECT 命令
func (c *StompConnection) handleDisconnect(frame *Frame) {
	// 发送收据（如果请求）
	if receipt := frame.GetHeader(HdrReceipt); receipt != "" {
		c.SendReceipt(receipt)
	}

	// 关闭连接
	c.cancel()
}

// SendFrame 发送 STOMP 帧
func (c *StompConnection) SendFrame(frame *Frame) {
	data, err := frame.Marshal()
	if err != nil {
		logger.Error("STOMP 帧序列化失败", zap.Error(err))
		return
	}

	select {
	case c.sendChan <- data:
	default:
		logger.Warn("发送通道已满，丢弃消息",
			zap.String("connectionId", c.connectionId),
		)
	}
}

// SendError 发送 ERROR 帧
func (c *StompConnection) SendError(message string) {
	frame := NewErrorFrame(message)
	c.SendFrame(frame)
}

// SendReceipt 发送 RECEIPT 帧
func (c *StompConnection) SendReceipt(receiptId string) {
	frame := NewReceiptFrame(receiptId)
	c.SendFrame(frame)
}

// Close 关闭连接
func (c *StompConnection) Close() {
	c.cancel()
	c.wsConn.Close()
	close(c.sendChan)
}

// generateConnectionId 生成连接 ID
func generateConnectionId() string {
	return fmt.Sprintf("conn-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
