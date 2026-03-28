package sse

import "time"

const (
	TopicDict        = "dict"
	TopicOnlineCount = "online-count"
	TopicSystem      = "system"
)

type DictChangeEvent struct {
	DictCode  string `json:"dictCode"`
	Timestamp int64  `json:"timestamp"`
}

func NewDictChangeEvent(dictCode string) *DictChangeEvent {
	return &DictChangeEvent{
		DictCode:  dictCode,
		Timestamp: time.Now().UnixMilli(),
	}
}

type OnlineUserDTO struct {
	Username     string `json:"username"`
	SessionCount int    `json:"sessionCount"`
	LoginTime    int64  `json:"loginTime"`
}
