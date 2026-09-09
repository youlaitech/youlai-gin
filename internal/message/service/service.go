package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	msgModel "youlai-gin/internal/message/model"
	"youlai-gin/internal/common/logger"
)

// SseEmitter 单个 SSE 连接，封装响应写入与关闭信号
type SseEmitter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// NewSseEmitter 基于 HTTP 响应创建 SSE 连接，要求底层支持流式写入
func NewSseEmitter(w http.ResponseWriter) (*SseEmitter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	return &SseEmitter{
		w:       w,
		flusher: flusher,
		done:    make(chan struct{}),
	}, nil
}

// Send 向客户端推送一条命名事件
func (e *SseEmitter) Send(eventName string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	fmt.Fprintf(e.w, "event: %s\n", eventName)
	fmt.Fprintf(e.w, "data: %s\n\n", jsonData)
	e.flusher.Flush()
	return nil
}

func (e *SseEmitter) Close() {
	close(e.done)
}

// Done 返回连接关闭信号，供调用方监听
func (e *SseEmitter) Done() <-chan struct{} {
	return e.done
}

// SendHeartbeat 发送注释行心跳以保活连接
func (e *SseEmitter) SendHeartbeat() error {
	select {
	case <-e.done:
		return fmt.Errorf("连接已关闭")
	default:
	}
	_, err := fmt.Fprintf(e.w, ": heartbeat\n\n")
	if err != nil {
		return err
	}
	e.flusher.Flush()
	return nil
}

// SseService 管理 SSE 连接与事件广播
type SseService struct {
	registry *SseSessionRegistry
}

// NewSseService 创建 SSE 服务
func NewSseService() *SseService {
	return &SseService{
		registry: NewSseSessionRegistry(),
	}
}

// CreateConnection 建立连接、登记会话并广播最新在线人数
func (s *SseService) CreateConnection(username string, w http.ResponseWriter) (*SseEmitter, error) {
	emitter, err := NewSseEmitter(w)
	if err != nil {
		return nil, err
	}

	s.registry.UserConnected(username, emitter)

	// 发送初始在线人数
	if err := emitter.Send(msgModel.TopicOnlineCount, s.registry.GetOnlineUserCount()); err != nil {
		logger.Warn("发送初始在线用户数失败", zap.Error(err))
	}

	logger.Info("SSE连接已建立", zap.String("username", username), zap.Int("online", s.registry.GetOnlineUserCount()))

	// 广播在线人数
	s.SendOnlineCount()

	return emitter, nil
}

// SendDictChange 广播字典变更事件
func (s *SseService) SendDictChange(dictCode string) {
	if dictCode == "" {
		return
	}
	event := msgModel.NewDictChangeEvent(dictCode)
	s.broadcast(msgModel.TopicDict, event)
	logger.Debug("字典变更通知已发送", zap.String("dictCode", dictCode))
}

func (s *SseService) SendOnlineCount() {
	count := s.registry.GetOnlineUserCount()
	s.broadcast(msgModel.TopicOnlineCount, count)
}

// SendToUser 向指定用户的全部连接推送事件
func (s *SseService) SendToUser(username string, eventName string, data interface{}) {
	emitters := s.registry.GetUserEmitters(username)
	if emitters == nil {
		return
	}
	for _, emitter := range emitters {
		if err := emitter.Send(eventName, data); err != nil {
			logger.Warn("发送SSE事件失败", zap.String("username", username), zap.Error(err))
			s.registry.RemoveEmitter(emitter)
		}
	}
	logger.Debug("SSE事件已发送给用户", zap.String("username", username), zap.String("event", eventName))
}

func (s *SseService) GetOnlineUsers() []*msgModel.OnlineUserDTO {
	return s.registry.GetOnlineUsers()
}

func (s *SseService) GetOnlineUserCount() int {
	return s.registry.GetOnlineUserCount()
}

// SendSystemMessage 广播系统通知
func (s *SseService) SendSystemMessage(message string) {
	systemMessage := map[string]interface{}{
		"sender":    "系统通知",
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}
	s.broadcast(msgModel.TopicSystem, systemMessage)
	logger.Debug("系统消息已发送", zap.String("message", message))
}

func (s *SseService) RemoveEmitter(emitter *SseEmitter) {
	s.registry.RemoveEmitter(emitter)
}

// CloseAll 关闭所有SSE连接，在服务关闭时调用
func (s *SseService) CloseAll() {
	s.registry.CloseAll()
}

func (s *SseService) broadcast(eventName string, data interface{}) {
	emitters := s.registry.GetAllEmitters()
	for _, emitter := range emitters {
		if err := emitter.Send(eventName, data); err != nil {
			s.registry.RemoveEmitter(emitter)
		}
	}
}

var defaultSseService *SseService

// InitSseService 初始化全局 SSE 服务实例
func InitSseService() {
	defaultSseService = NewSseService()
}

// GetSseService 返回全局 SSE 服务实例
func GetSseService() *SseService {
	return defaultSseService
}