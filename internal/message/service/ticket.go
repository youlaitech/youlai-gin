package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"youlai-gin/internal/common/redis"
)

// 票据有效期较短，取到后需尽快完成握手
const (
	ticketTTL       = 30 * time.Second
	ticketKeyPrefix = "realtime:ticket:"
)

// TicketGrant 票据承载的握手授权内容，不得写入响应或日志
type TicketGrant struct {
	AccessToken      string   `json:"accessToken"`
	Username         string   `json:"username"`
	ClientInstanceID string   `json:"clientInstanceId"`
	Platform         string   `json:"platform"`
	Topics           []string `json:"topics"`
	ExpiresAt        int64    `json:"expiresAt"`
}

// CreateTicket 校验主题后签发一次性票据
func (s *RealtimeService) CreateTicket(username string, accessToken string, topics []string) (string, error) {
	ticket := strings.ReplaceAll(uuid.NewString(), "-", "")
	grant := TicketGrant{
		AccessToken: accessToken,
		Username:    username,
		Topics:      topics,
		ExpiresAt:   time.Now().Add(ticketTTL).UnixMilli(),
	}
	data, err := json.Marshal(grant)
	if err != nil {
		return "", err
	}
	if err := redis.Client.Set(context.Background(), ticketKey(ticket), data, ticketTTL).Err(); err != nil {
		return "", err
	}
	return ticket, nil
}

// ConsumeTicket 原子取出并删除票据，票据无效、已使用或已过期返回错误
func (s *RealtimeService) ConsumeTicket(ticket string) (*TicketGrant, error) {
	if ticket == "" {
		return nil, errors.New("票据无效")
	}
	raw, err := redis.Client.GetDel(context.Background(), ticketKey(ticket)).Result()
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, errors.New("票据无效或已使用")
	}
	var grant TicketGrant
	if err := json.Unmarshal([]byte(raw), &grant); err != nil {
		return nil, errors.New("票据内容解析失败")
	}
	if grant.ExpiresAt <= time.Now().UnixMilli() {
		return nil, errors.New("票据已过期")
	}
	return &grant, nil
}

// ticketKey 构造存储键，Redis 键只保存票据摘要
func ticketKey(ticket string) string {
	digest := sha256.Sum256([]byte(ticket))
	return ticketKeyPrefix + hex.EncodeToString(digest[:])
}
