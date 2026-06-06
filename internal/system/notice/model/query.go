package model

import common "youlai-gin/pkg/model"

// NoticeQuery 通知分页查询
type NoticeQuery struct {
	common.BaseQuery
	Title  string `form:"title"` // 标题

	Type   *int   `form:"type"` // 类型

	Status *int   `form:"status"` // 状态

}

// UserNoticeQuery 用户通知查询
type UserNoticeQuery struct {
	common.BaseQuery
	Type   *int `form:"type"` // 类型

	IsRead *int `form:"isRead"` // 0:未读 1:已读
}
