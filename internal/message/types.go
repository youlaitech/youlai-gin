// Package message 定义 SSE 推送的消息主题与数据结构。
package message

import "time"

// SSE 消息主题
const (
	TopicDict        = "dict"         // 字典数据变更
	TopicOnlineCount = "online-count" // 在线人数变化
	TopicSystem      = "system"       // 系统级通知
)

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