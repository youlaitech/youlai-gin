package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RealtimeMessage 实时消息，与 youlai-boot 的协议保持一致
type RealtimeMessage struct {
	Version int         `json:"version"`
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// NewRealtimeMessage 构造服务端消息并生成唯一标识
func NewRealtimeMessage(msgType string, payload interface{}) *RealtimeMessage {
	return &RealtimeMessage{
		Version: 1,
		ID:      uuid.NewString(),
		Type:    msgType,
		Payload: payload,
	}
}

// PongPayload 客户端心跳响应载荷
type PongPayload struct {
	PingID string `json:"pingId"`
}

// ParsePong 解析客户端心跳响应，非心跳消息返回 false
func ParsePong(data []byte) (string, bool) {
	var raw struct {
		Version int          `json:"version"`
		Type    string       `json:"type"`
		Payload *PongPayload `json:"payload"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", false
	}
	if raw.Version != 1 || raw.Type != EventHeartbeatPong || raw.Payload == nil {
		return "", false
	}
	if raw.Payload.PingID == "" {
		return "", false
	}
	return raw.Payload.PingID, true
}

// RealtimeMessageEnvelope 客户端上行消息，用于提取心跳标识
type RealtimeMessageEnvelope struct {
	Version int          `json:"version"`
	ID      string       `json:"id"`
	Type    string       `json:"type"`
	Payload *PongPayload `json:"payload"`
}

// NowMillis 当前毫秒时间戳
func NowMillis() int64 {
	return time.Now().UnixMilli()
}
