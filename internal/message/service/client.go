package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	msgModel "youlai-gin/internal/message/model"
)

// RealtimeClient 单个 WebSocket 连接，负责串行化写入与关闭
type RealtimeClient struct {
	conn          *websocket.Conn
	send          chan []byte
	username      string
	topics        map[string]struct{}
	done          chan struct{}
	closeOnce     sync.Once
	mu            sync.Mutex
	pendingPingID string
	lastPong      time.Time
}

// NewRealtimeClient 创建连接客户端
func NewRealtimeClient(conn *websocket.Conn, username string, topics map[string]struct{}) *RealtimeClient {
	return &RealtimeClient{
		conn:     conn,
		send:     make(chan []byte, 64),
		username: username,
		topics:   topics,
		done:     make(chan struct{}),
		lastPong: time.Now(),
	}
}

// Username 返回所属用户名
func (c *RealtimeClient) Username() string {
	return c.username
}

// HasTopic 判断是否订阅指定主题
func (c *RealtimeClient) HasTopic(topic string) bool {
	_, ok := c.topics[topic]
	return ok
}

// Send 队列化一条消息，写入由单独协程完成
func (c *RealtimeClient) Send(msg *msgModel.RealtimeMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	case <-c.done:
		return fmt.Errorf("连接已关闭")
	}
}

// Done 返回连接关闭信号
func (c *RealtimeClient) Done() <-chan struct{} {
	return c.done
}

// Close 关闭连接，只生效一次
func (c *RealtimeClient) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}

// SendChan 返回发送队列，写入协程从中取数据发出
func (c *RealtimeClient) SendChan() <-chan []byte {
	return c.send
}

// ConfirmPong 确认心跳响应，仅接受当前待确认的心跳
func (c *RealtimeClient) ConfirmPong(pingID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pendingPingID != pingID {
		return
	}
	c.pendingPingID = ""
	c.lastPong = time.Now()
}

// ExpectPong 记录待确认的心跳标识
func (c *RealtimeClient) ExpectPong(pingID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pendingPingID = pingID
}

// LastPongAt 返回最近一次心跳响应时间
func (c *RealtimeClient) LastPongAt() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastPong
}
