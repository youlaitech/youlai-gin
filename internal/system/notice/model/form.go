package model

import "youlai-gin/pkg/types"

// NoticeForm 通知公告表单
// Status 对应表字段 publish_status（名称不一致），用 gorm:"column:publish_status" 显式映射。
// PublishTime/TargetUsers 需要 service 层做格式/类型转换，用 gorm:"-" 排除，避免 BuildPatchMap 直接写入。
type NoticeForm struct {
	ID          types.BigInt   `json:"id"` // 主键
	Title       string         `json:"title" binding:"required"`
	Content     string         `json:"content" binding:"required"`
	Type        int            `json:"type"`                                // 类型
	Level       string         `json:"level"`                               // L:普通 M:中等 H:重要
	Status      int            `json:"status" gorm:"column:publish_status"` // 状态(1发布0草稿-1撤回)
	PublishTime string         `json:"publishTime" gorm:"-"`                // 由 service 解析后写入 publish_time
	TargetType  int            `json:"targetType"`
	TargetUsers []types.BigInt `json:"targetUsers" gorm:"-"` // 目标用户，转成 target_user_ids 逗号串
}
