package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

// NoticeForm 通知公告表单
type NoticeForm struct {
	ID          types.BigInt   `json:"id"`
	Title       string         `json:"title" binding:"required"`
	Content     string         `json:"content" binding:"required"`
	Type        int            `json:"type"`
	Level       string         `json:"level"` // L:普通 M:中等 H:重要
	Status      int            `json:"status"`
	PublishTime string         `json:"publishTime"`
	TargetType  int            `json:"targetType"`
	TargetUsers []types.BigInt `json:"targetUsers"` // 目标用户ID列表
}

// NoticePageQuery 通知分页查询
type NoticePageQuery struct {
	common.PageQuery
	Title  string `form:"title"`
	Type   *int   `form:"type"`
	Status *int   `form:"status"`
}

// UserNoticeQuery 用户通知查询
type UserNoticeQuery struct {
	common.PageQuery
	Type   *int `form:"type"`
	IsRead *int `form:"isRead"` // 0:未读 1:已读
}
