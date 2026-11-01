package service

import (
	"sync"
	"time"

	"github.com/google/uuid"

	msgModel "youlai-gin/internal/message/model"
)

// RealtimeSessionRegistry 管理实时连接会话，支持多设备同时登录
type RealtimeSessionRegistry struct {
	mu          sync.RWMutex
	userClients map[string]map[*RealtimeClient]struct{}
	clientTime  map[*RealtimeClient]time.Time
	nodeId      string
}

// NewRealtimeSessionRegistry 创建会话注册表
func NewRealtimeSessionRegistry() *RealtimeSessionRegistry {
	return &RealtimeSessionRegistry{
		userClients: make(map[string]map[*RealtimeClient]struct{}),
		clientTime:  make(map[*RealtimeClient]time.Time),
		nodeId:      uuid.NewString(),
	}
}

// UserConnected 登记新连接
func (r *RealtimeSessionRegistry) UserConnected(username string, client *RealtimeClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	clients, ok := r.userClients[username]
	if !ok {
		clients = make(map[*RealtimeClient]struct{})
		r.userClients[username] = clients
	}
	clients[client] = struct{}{}
	r.clientTime[client] = time.Now()
}

// RemoveClient 移除连接，用户无连接时清空记录
func (r *RealtimeSessionRegistry) RemoveClient(client *RealtimeClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clientTime, client)
	clients, ok := r.userClients[client.Username()]
	if !ok {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(r.userClients, client.Username())
	}
}

// GetUserClients 返回指定用户的全部连接
func (r *RealtimeSessionRegistry) GetUserClients(username string) []*RealtimeClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := r.userClients[username]
	result := make([]*RealtimeClient, 0, len(clients))
	for client := range clients {
		result = append(result, client)
	}
	return result
}

// GetAllClients 返回全部连接
func (r *RealtimeSessionRegistry) GetAllClients() []*RealtimeClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*RealtimeClient, 0, len(r.clientTime))
	for client := range r.clientTime {
		result = append(result, client)
	}
	return result
}

// GetOnlineUserCount 在线用户数，按用户去重
func (r *RealtimeSessionRegistry) GetOnlineUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.userClients)
}

// GetOnlineConnectionCount 在线连接总数，多标签页会重复计数
func (r *RealtimeSessionRegistry) GetOnlineConnectionCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clientTime)
}

// NodeId 节点标识，进程内固定
func (r *RealtimeSessionRegistry) NodeId() string {
	return r.nodeId
}

// GetOnlineUsers 在线用户列表，登录时间取该用户最早的连接时间
func (r *RealtimeSessionRegistry) GetOnlineUsers() []*msgModel.OnlineUserDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*msgModel.OnlineUserDTO, 0, len(r.userClients))
	for username, clients := range r.userClients {
		earliest := time.Now()
		for client := range clients {
			if t, ok := r.clientTime[client]; ok && t.Before(earliest) {
				earliest = t
			}
		}
		result = append(result, &msgModel.OnlineUserDTO{
			Username:     username,
			SessionCount: len(clients),
			LoginTime:    earliest.UnixMilli(),
		})
	}
	return result
}

// CloseAll 关闭全部连接
func (r *RealtimeSessionRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for client := range r.clientTime {
		client.Close()
	}
	r.userClients = make(map[string]map[*RealtimeClient]struct{})
	r.clientTime = make(map[*RealtimeClient]time.Time)
}
