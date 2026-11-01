// Package model 定义实时推送的主题常量与数据结构。
package model

import "time"

// 订阅主题
const (
	TopicNotice   = "notice"   // 通知公告
	TopicDict     = "dict"     // 字典数据变更
	TopicPresence = "presence" // 在线状态
)

// 事件类型
const (
	EventConnectionReady  = "connection.ready"  // 连接就绪，data 为 heartbeatTimeoutMs
	EventHeartbeatPing    = "heartbeat.ping"    // 服务端心跳，data 为空对象
	EventNoticeChanged    = "notice.changed"    // 公告变更，data 为 noticeId 与 action
	EventDictChanged      = "dict.changed"      // 字典变更，data 为 dictCode
	EventPresenceSnapshot = "presence.snapshot" // 在线状态快照，data 为在线数与节点信息
	EventHeartbeatPong    = "heartbeat.pong"    // 客户端心跳响应
)

// HeartbeatTimeoutMs 心跳应答超时时间（毫秒），与前端协议保持一致
const HeartbeatTimeoutMs = 60000

// SupportedTopics 当前支持的订阅主题，订阅须包含 notice
var SupportedTopics = map[string]struct{}{
	TopicNotice:   {},
	TopicDict:     {},
	TopicPresence: {},
}

// DictChangeEvent 字典变更事件，携带变更的字典编码与毫秒时间戳
type DictChangeEvent struct {
	DictCode  string `json:"dictCode"`
	Timestamp int64  `json:"timestamp"`
}

// NewDictChangeEvent 按当前时间构造字典变更事件
func NewDictChangeEvent(dictCode string) *DictChangeEvent {
	return &DictChangeEvent{
		DictCode:  dictCode,
		Timestamp: time.Now().UnixMilli(),
	}
}

// OnlineUserDTO 在线用户信息，含用户名、会话数与登录时间戳
type OnlineUserDTO struct {
	Username     string `json:"username"`
	SessionCount int    `json:"sessionCount"`
	LoginTime    int64  `json:"loginTime"`
}
