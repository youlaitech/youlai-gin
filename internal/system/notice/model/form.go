package model

import "youlai-gin/pkg/types"

// NoticeForm 通知公告表单
type NoticeForm struct {
	ID          types.BigInt   `json:"id"` // 主键
	Title       string         `json:"title" binding:"required"`
	Content     string         `json:"content" binding:"required"`
	Type        int            `json:"type"` // 类型(C目录/M菜单/B按钮)
	Level       string         `json:"level"` // L:普通 M:中等 H:重要
	Status      int            `json:"status"` // 状态(1启用0禁用)
	PublishTime string         `json:"publishTime"`
	TargetType  int            `json:"targetType"`
	TargetUsers []types.BigInt `json:"targetUsers"` // 目标用户ID列表
}
