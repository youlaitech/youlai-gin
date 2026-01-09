package service

import (
	"encoding/json"
	"time"

	"youlai-gin/internal/system/notice/model"
	"youlai-gin/internal/system/notice/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/websocket"
)

// GetNoticePage 通知分页查询
func GetNoticePage(query *model.NoticeQuery) (*common.PagedData, error) {
	list, total, err := repository.GetNoticePage(query)
	if err != nil {
		return nil, errs.SystemError("查询通知列表失败")
	}

	pageMeta := common.NewPageMeta(query.PageNum, query.PageSize, total)
	return &common.PagedData{Data: list, Page: pageMeta}, nil
}

// GetNoticeByID 根据ID获取通知
func GetNoticeByID(id int64) (*model.Notice, error) {
	return repository.GetNoticeByID(id)
}

// SaveNotice 保存通知（新增或更新）
func SaveNotice(form *model.NoticeForm) error {
	notice := &model.Notice{
		ID:          form.ID,
		Title:       form.Title,
		Content:     form.Content,
		Type:        form.Type,
		Level:       form.Level,
		Status:      form.Status,
		PublishTime: form.PublishTime,
		TargetType:  form.TargetType,
	}

	// 转换目标用户列表为JSON
	if len(form.TargetUsers) > 0 {
		targetUsersJSON, _ := json.Marshal(form.TargetUsers)
		notice.TargetUsers = string(targetUsersJSON)
	}

	// 如果没有设置发布时间，使用当前时间
	if notice.PublishTime == "" && notice.Status == 1 {
		notice.PublishTime = time.Now().Format("2006-01-02 15:04:05")
	}

	var err error
	if notice.ID > 0 {
		err = repository.UpdateNotice(notice)
	} else {
		err = repository.CreateNotice(notice)
	}

	if err != nil {
		return errs.SystemError("保存通知失败")
	}

	// 如果是发布状态，推送WebSocket消息
	if notice.Status == 1 {
		go pushNotice(notice, form.TargetUsers)
	}

	return nil
}

// DeleteNotice 删除通知
func DeleteNotice(id int64) error {
	if err := repository.DeleteNotice(id); err != nil {
		return errs.SystemError("删除通知失败")
	}
	return nil
}

// GetUserNoticePage 获取用户通知列表
func GetUserNoticePage(userID int64, query *model.UserNoticeQuery) (*common.PagedData, error) {
	list, total, err := repository.GetUserNoticePage(userID, query)
	if err != nil {
		return nil, errs.SystemError("查询用户通知列表失败")
	}

	pageMeta := common.NewPageMeta(query.PageNum, query.PageSize, total)
	return &common.PagedData{Data: list, Page: pageMeta}, nil
}

// MarkNoticeAsRead 标记通知为已读
func MarkNoticeAsRead(noticeID, userID int64) error {
	if err := repository.MarkNoticeAsRead(noticeID, userID); err != nil {
		return errs.SystemError("标记通知已读失败")
	}
	return nil
}

// GetUnreadCount 获取未读通知数量
func GetUnreadCount(userID int64) (int64, error) {
	return repository.GetUnreadCount(userID)
}

// pushNotice 推送通知（WebSocket）
func pushNotice(notice *model.Notice, targetUsers []types.BigInt) {
	if websocket.DefaultHub == nil {
		return
	}

	message := &websocket.Message{
		Type:    "notice",
		Title:   notice.Title,
		Content: notice.Content,
		Data: map[string]interface{}{
			"id":    notice.ID,
			"type":  notice.Type,
			"level": notice.Level,
		},
	}

	if notice.TargetType == 1 {
		// 广播给所有在线用户
		websocket.DefaultHub.BroadcastMessage(message)
	} else if len(targetUsers) > 0 {
		// 发送给指定用户
		// websocket.SendMessage expects []int64, convert back
		ids := make([]int64, len(targetUsers))
		for i, id := range targetUsers {
			ids[i] = int64(id)
		}
		websocket.DefaultHub.SendMessage(ids, message)
	}
}

// PublishNotice 发布通知
func PublishNotice(id int64) error {
	notice, err := repository.GetNoticeByID(id)
	if err != nil {
		return errs.NotFound("通知不存在")
	}

	notice.Status = 1
	notice.PublishTime = time.Now().Format("2006-01-02 15:04:05")

	if err := repository.UpdateNotice(notice); err != nil {
		return errs.SystemError("发布通知失败")
	}

	// 推送WebSocket消息
	var targetUsers []types.BigInt
	if notice.TargetUsers != "" {
		json.Unmarshal([]byte(notice.TargetUsers), &targetUsers)
	}
	go pushNotice(notice, targetUsers)

	return nil
}

// RevokeNotice 撤回通知
func RevokeNotice(id int64) error {
	notice, err := repository.GetNoticeByID(id)
	if err != nil {
		return errs.NotFound("通知不存在")
	}

	if notice.Status != 1 {
		return errs.BadRequest("通知未发布或已撤回")
	}

	notice.Status = 0 // 设置为草稿状态

	if err := repository.UpdateNotice(notice); err != nil {
		return errs.SystemError("撤回通知失败")
	}

	return nil
}
