package sse

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"youlai-gin/pkg/logger"
)

type SessionInfo struct {
	Username    string
	ConnectTime int64
}

type SseSessionRegistry struct {
	mu               sync.RWMutex
	userEmittersMap  map[string]map[*SseEmitter]bool
	emitterUserMap   map[*SseEmitter]*SessionInfo
	emitterTimeMap   map[*SseEmitter]int64
}

func NewSseSessionRegistry() *SseSessionRegistry {
	return &SseSessionRegistry{
		userEmittersMap: make(map[string]map[*SseEmitter]bool),
		emitterUserMap:  make(map[*SseEmitter]*SessionInfo),
		emitterTimeMap:  make(map[*SseEmitter]int64),
	}
}

func (r *SseSessionRegistry) UserConnected(username string, emitter *SseEmitter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.userEmittersMap[username] == nil {
		r.userEmittersMap[username] = make(map[*SseEmitter]bool)
	}
	r.userEmittersMap[username][emitter] = true
	r.emitterUserMap[emitter] = &SessionInfo{
		Username:    username,
		ConnectTime: time.Now().UnixMilli(),
	}
	r.emitterTimeMap[emitter] = time.Now().UnixMilli()

	logger.Debug("SSE连接已建立", zap.String("username", username), zap.Int("online", len(r.userEmittersMap)))
}

func (r *SseSessionRegistry) RemoveEmitter(emitter *SseEmitter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionInfo, ok := r.emitterUserMap[emitter]
	if !ok {
		return
	}

	delete(r.emitterUserMap, emitter)
	delete(r.emitterTimeMap, emitter)

	emitters, ok := r.userEmittersMap[sessionInfo.Username]
	if ok {
		delete(emitters, emitter)
		if len(emitters) == 0 {
			delete(r.userEmittersMap, sessionInfo.Username)
			logger.Debug("用户所有SSE连接已断开", zap.String("username", sessionInfo.Username))
		}
	}
}

func (r *SseSessionRegistry) GetOnlineUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.userEmittersMap)
}

func (r *SseSessionRegistry) GetTotalConnectionCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.emitterUserMap)
}

func (r *SseSessionRegistry) GetOnlineUsers() []*OnlineUserDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*OnlineUserDTO, 0, len(r.userEmittersMap))
	for username, emitters := range r.userEmittersMap {
		var earliestLoginTime int64 = -1
		for emitter := range emitters {
			if t, ok := r.emitterTimeMap[emitter]; ok {
				if earliestLoginTime == -1 || t < earliestLoginTime {
					earliestLoginTime = t
				}
			}
		}
		if earliestLoginTime == -1 {
			earliestLoginTime = time.Now().UnixMilli()
		}
		result = append(result, &OnlineUserDTO{
			Username:     username,
			SessionCount: len(emitters),
			LoginTime:    earliestLoginTime,
		})
	}
	return result
}

func (r *SseSessionRegistry) GetAllEmitters() []*SseEmitter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emitters := make([]*SseEmitter, 0, len(r.emitterUserMap))
	for emitter := range r.emitterUserMap {
		emitters = append(emitters, emitter)
	}
	return emitters
}

func (r *SseSessionRegistry) GetUserEmitters(username string) []*SseEmitter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emitterSet, ok := r.userEmittersMap[username]
	if !ok {
		return nil
	}
	emitters := make([]*SseEmitter, 0, len(emitterSet))
	for emitter := range emitterSet {
		emitters = append(emitters, emitter)
	}
	return emitters
}

func (r *SseSessionRegistry) IsUserOnline(username string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	emitters, ok := r.userEmittersMap[username]
	return ok && len(emitters) > 0
}

// CloseAll 关闭所有SSE连接，在服务关闭时调用
func (r *SseSessionRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.emitterUserMap)
	if count == 0 {
		return
	}
	logger.Info("应用关闭，主动断开SSE连接...", zap.Int("count", count))

	for emitter := range r.emitterUserMap {
		if emitter.done != nil {
			select {
			case <-emitter.done:
			default:
				close(emitter.done)
			}
		}
	}
	r.userEmittersMap = make(map[string]map[*SseEmitter]bool)
	r.emitterUserMap = make(map[*SseEmitter]*SessionInfo)
	r.emitterTimeMap = make(map[*SseEmitter]int64)

	logger.Info("所有SSE连接已断开")
}
