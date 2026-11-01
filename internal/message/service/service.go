package service

import (
	"time"

	"go.uber.org/zap"

	"youlai-gin/internal/common/logger"
	msgModel "youlai-gin/internal/message/model"
)

// RealtimeService 管理实时连接与消息广播
type RealtimeService struct {
	registry *RealtimeSessionRegistry
}

// NewRealtimeService 创建实时服务
func NewRealtimeService() *RealtimeService {
	return &RealtimeService{registry: NewRealtimeSessionRegistry()}
}

// RegisterClient 登记连接
func (s *RealtimeService) RegisterClient(client *RealtimeClient) {
	s.registry.UserConnected(client.Username(), client)
}

// RemoveClient 移除连接
func (s *RealtimeService) RemoveClient(client *RealtimeClient) {
	s.registry.RemoveClient(client)
}

// Broadcast 向订阅了该主题的连接广播消息
func (s *RealtimeService) Broadcast(topic string, msgType string, payload interface{}) {
	message := msgModel.NewRealtimeMessage(msgType, payload)
	for _, client := range s.registry.GetAllClients() {
		if !client.HasTopic(topic) {
			continue
		}
		if err := client.Send(message); err != nil {
			logger.Warn("发送实时消息失败", zap.String("username", client.Username()), zap.Error(err))
			s.registry.RemoveClient(client)
		}
	}
}

// SendToUser 向指定用户的全部连接推送消息
func (s *RealtimeService) SendToUser(username string, msgType string, payload interface{}) {
	message := msgModel.NewRealtimeMessage(msgType, payload)
	for _, client := range s.registry.GetUserClients(username) {
		if err := client.Send(message); err != nil {
			logger.Warn("发送实时消息失败", zap.String("username", username), zap.Error(err))
			s.registry.RemoveClient(client)
		}
	}
}

// SendDictChange 广播字典变更
func (s *RealtimeService) SendDictChange(dictCode string) {
	if dictCode == "" {
		return
	}
	s.Broadcast(msgModel.TopicDict, msgModel.EventDictChanged, map[string]any{"dictCode": dictCode})
}

// SendOnlineCount 广播在线状态快照
func (s *RealtimeService) SendOnlineCount() {
	s.Broadcast(msgModel.TopicPresence, msgModel.EventPresenceSnapshot, s.PresenceSnapshot())
}

// GetOnlineUsers 在线用户列表
func (s *RealtimeService) GetOnlineUsers() []*msgModel.OnlineUserDTO {
	return s.registry.GetOnlineUsers()
}

// GetOnlineUserCount 在线用户数
func (s *RealtimeService) GetOnlineUserCount() int {
	return s.registry.GetOnlineUserCount()
}

// GetOnlineConnectionCount 在线连接总数
func (s *RealtimeService) GetOnlineConnectionCount() int {
	return s.registry.GetOnlineConnectionCount()
}

// PresenceSnapshot 生成在线状态快照
func (s *RealtimeService) PresenceSnapshot() map[string]any {
	return map[string]any{
		"onlineUsers":       s.registry.GetOnlineUserCount(),
		"onlineConnections": s.registry.GetOnlineConnectionCount(),
		"nodeId":            s.registry.NodeId(),
		"observedAt":        time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}
}

// CloseAll 关闭全部连接
func (s *RealtimeService) CloseAll() {
	s.registry.CloseAll()
}

var defaultRealtimeService *RealtimeService

// InitRealtimeService 初始化全局实时服务实例
func InitRealtimeService() {
	defaultRealtimeService = NewRealtimeService()
}

// GetRealtimeService 返回全局实时服务实例
func GetRealtimeService() *RealtimeService {
	return defaultRealtimeService
}
