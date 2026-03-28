package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"youlai-gin/pkg/logger"
)

type SseEmitter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

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

func (e *SseEmitter) Done() <-chan struct{} {
	return e.done
}

func (e *SseEmitter) SendHeartbeat() {
	fmt.Fprintf(e.w, ": heartbeat\n\n")
	e.flusher.Flush()
}

type SseService struct {
	registry *SseSessionRegistry
}

func NewSseService() *SseService {
	return &SseService{
		registry: NewSseSessionRegistry(),
	}
}

func (s *SseService) CreateConnection(username string, w http.ResponseWriter) (*SseEmitter, error) {
	emitter, err := NewSseEmitter(w)
	if err != nil {
		return nil, err
	}

	s.registry.UserConnected(username, emitter)

	// Send initial online count
	if err := emitter.Send(TopicOnlineCount, s.registry.GetOnlineUserCount()); err != nil {
		logger.Warn("发送初始在线用户数失败", zap.Error(err))
	}

	logger.Info("SSE连接已建立", zap.String("username", username), zap.Int("online", s.registry.GetOnlineUserCount()))

	// Broadcast online count to all
	s.SendOnlineCount()

	return emitter, nil
}

func (s *SseService) SendDictChange(dictCode string) {
	if dictCode == "" {
		return
	}
	event := NewDictChangeEvent(dictCode)
	s.broadcast(TopicDict, event)
	logger.Debug("字典变更通知已发送", zap.String("dictCode", dictCode))
}

func (s *SseService) SendOnlineCount() {
	count := s.registry.GetOnlineUserCount()
	s.broadcast(TopicOnlineCount, count)
}

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

func (s *SseService) GetOnlineUsers() []*OnlineUserDTO {
	return s.registry.GetOnlineUsers()
}

func (s *SseService) GetOnlineUserCount() int {
	return s.registry.GetOnlineUserCount()
}

func (s *SseService) SendSystemMessage(message string) {
	systemMessage := map[string]interface{}{
		"sender":    "系统通知",
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}
	s.broadcast(TopicSystem, systemMessage)
	logger.Debug("系统消息已发送", zap.String("message", message))
}

func (s *SseService) RemoveEmitter(emitter *SseEmitter) {
	s.registry.RemoveEmitter(emitter)
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

func InitSseService() {
	defaultSseService = NewSseService()
}

func GetSseService() *SseService {
	return defaultSseService
}
