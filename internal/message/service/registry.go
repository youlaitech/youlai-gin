package service

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"youlai-gin/internal/common/logger"
	msgModel "youlai-gin/internal/message/model"
)

// SessionInfo 单条 SSE 会话信息
type SessionInfo struct {
	Username    string
	ConnectTime int64
}

// SseSessionRegistry 维护「用户 ↔ 连接」映射，支撑在线人数统计与定向推送
// 内部实现（非 DB 仓储）：连接登记写在此层，SseService 只感知会话概念
type SseSessionRegistry struct {
	mu              sync.RWMutex
	userEmittersMap map[string]map[*SseEmitter]bool
	emitterUserMap  map[*SseEmitter]*SessionInfo
	emitterTimeMap  map[*SseEmitter]int64
}

func NewSseSessionRegistry() *SseSessionRegistry {
	return &SseSessionRegistry{
		userEmittersMap: make(map[string]map[*SseEmitter]bool),
		emitterUserMap:  make(map[*SseEmitter]*SessionInfo),
		emitterTimeMap:  make(map[*SseEmitter]int64),
	}
}

// UserConnected 登记用户的一条新连接
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

// RemoveEmitter 移除连接，用户连接全部断开时一并清理其映射
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

// GetOnlineUserCount 在线用户数（按用户名去重）
func (r *SseSessionRegistry) GetOnlineUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.userEmittersMap)
}

// GetTotalConnectionCount 连接总数（含同一用户的多个标签页）
func (r *SseSessionRegistry) GetTotalConnectionCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.emitterUserMap)
}

// GetOnlineUsers 在线用户列表，含各自会话数与最早登录时间
func (r *SseSessionRegistry) GetOnlineUsers() []*msgModel.OnlineUserDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*msgModel.OnlineUserDTO, 0, len(r.userEmittersMap))
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
		result = append(result, &msgModel.OnlineUserDTO{
			Username:     username,
			SessionCount: len(emitters),
			LoginTime:    earliestLoginTime,
		})
	}
	return result
}

// GetAllEmitters 返回全部活跃连接
func (r *SseSessionRegistry) GetAllEmitters() []*SseEmitter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emitters := make([]*SseEmitter, 0, len(r.emitterUserMap))
	for emitter := range r.emitterUserMap {
		emitters = append(emitters, emitter)
	}
	return emitters
}

// GetUserEmitters 返回某用户的全部连接
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

// IsUserOnline 判断用户是否有活跃连接
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
