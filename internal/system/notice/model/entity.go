package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

// Notice 通知公告实体
type Notice struct {
	ID          types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string       `gorm:"column:title;size:200;not null" json:"title"`
	Content     string       `gorm:"column:content;type:text" json:"content"`
	Type        int          `gorm:"column:type;default:1" json:"type"`             // 1:通知 2:公告
	Level       string       `gorm:"column:level;default:L" json:"level"`           // L:普通 M:中等 H:重要
	Status      int          `gorm:"column:publish_status;default:0" json:"publishStatus"` // 0:草稿 1:已发布 -1:已撤回
	PublishTime types.LocalTime `gorm:"column:publish_time" json:"publishTime"`
	TargetType  int          `gorm:"column:target_type;default:1" json:"targetType"`              // 1:全部 2:指定用户
	TargetUsers string       `gorm:"column:target_user_ids;type:varchar(255)" json:"targetUsers"` // 目标用户ID集合（逗号分隔）
	PublisherID types.BigInt `gorm:"column:publisher_id" json:"publisherId"`                      // 发布人ID
	PublisherName string     `gorm:"column:publisher_name;->" json:"publisherName"`
	RevokeTime  *types.LocalTime `gorm:"column:revoke_time" json:"revokeTime"`                        // 撤回时间
	common.BaseEntity
}

func (Notice) TableName() string {
	return "sys_notice"
}

// UserNotice 用户通知记录
type UserNotice struct {
	ID         types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	NoticeID   types.BigInt `gorm:"column:notice_id;not null" json:"noticeId"`
	UserID     types.BigInt `gorm:"column:user_id;not null" json:"userId"`
	IsRead     int          `gorm:"column:is_read;default:0" json:"isRead"` // 0:未读 1:已读
	ReadTime   string       `gorm:"column:read_time" json:"readTime"`
	CreateTime string       `gorm:"column:create_time" json:"createTime"`
	UpdateTime string       `gorm:"column:update_time" json:"updateTime"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (UserNotice) TableName() string {
	return "sys_user_notice"
}
