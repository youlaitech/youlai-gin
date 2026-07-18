package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/system/notice/model"
	"youlai-gin/internal/system/notice/repository"
	common "youlai-gin/pkg/model"
	"youlai-gin/internal/common/database"
	"youlai-gin/pkg/errs"
	"youlai-gin/internal/message"
	"youlai-gin/pkg/gormx"
	"youlai-gin/pkg/types"
)



// GetNoticePage 通知分页查询
func GetNoticePage(query *model.NoticeQuery) (*common.PagedData, error) {
	list, total, err := repository.GetNoticePage(query)
	if err != nil {
		return nil, errs.SystemError("查询通知列表失败")
	}

	return &common.PagedData{List: list, Total: total}, nil
}

// GetNoticeByID 根据ID获取通知
func GetNoticeByID(id int64) (*model.Notice, error) {
	return repository.GetNoticeByID(id)
}

// SaveNotice 保存通知（新增或更新）
// 新增时由表单构造完整实体；更新时由 BuildPatchMap(form) 生成部分更新映射，并补充 publish_time、
// target_user_ids 等需要转换的字段，从根上解决 GORM Updates(struct) 默认跳过零值的问题。
func SaveNotice(c *gin.Context, form *model.NoticeForm) error {
	ctx := appContext.OperatorCtx(c)

	parsePublishTime := func(s string) (*types.LocalTime, error) {
		if strings.TrimSpace(s) == "" {
			return nil, nil
		}
		parsed, err := time.ParseInLocation(types.TimeFormat, s, time.Local)
		if err != nil {
			return nil, err
		}
		t := types.LocalTime(parsed)
		return &t, nil
	}

	status := form.Status
	patch := gormx.BuildPatchMap(form)

	// publish_time：字符串时间转 *types.LocalTime 后写入
	if pt, err := parsePublishTime(form.PublishTime); err != nil {
		return errs.BadRequest("发布时间格式错误")
	} else if pt != nil {
		patch["publish_time"] = pt
	} else if v, ok := patch["publish_status"].(int); ok && v == 1 {
		// 发布且未指定时间时默认当前时间
		now := types.Now()
		patch["publish_time"] = &now
	}

	// target_user_ids：切片转 JSON 字符串后写入
	if len(form.TargetUsers) > 0 {
		if b, err := json.Marshal(form.TargetUsers); err == nil {
			patch["target_user_ids"] = string(b)
		}
	}

	if form.ID == 0 {
		notice := &model.Notice{
			Title:      form.Title,
			Content:    form.Content,
			Type:       form.Type,
			Level:      form.Level,
			Status:     status,
			TargetType: form.TargetType,
		}
		if pt, _ := parsePublishTime(form.PublishTime); pt != nil {
			notice.PublishTime = pt
		} else if status == 1 {
			now := types.Now()
			notice.PublishTime = &now
		}
		if len(form.TargetUsers) > 0 {
			if b, err := json.Marshal(form.TargetUsers); err == nil {
				notice.TargetUsers = string(b)
			}
		}
		if err := repository.CreateNotice(ctx, notice); err != nil {
			return errs.SystemError("创建通知失败")
		}
	} else {
		if err := repository.UpdateNotice(ctx, int64(form.ID), patch); err != nil {
			return errs.SystemError("更新通知失败")
		}
	}

	// 如果是发布状态，推送通知
	if status == 1 {
		notice := &model.Notice{ID: form.ID, Title: form.Title, Type: form.Type, Level: form.Level, Status: status}
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

	return &common.PagedData{List: list, Total: total}, nil
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

// pushNotice 推送通知（SSE）
func pushNotice(notice *model.Notice, targetUsers []types.BigInt) {
	sseService := message.GetSseService()
	if sseService == nil {
		return
	}

	noticeData := map[string]interface{}{
		"id":          notice.ID,
		"title":       notice.Title,
		"type":        notice.Type,
		"level":       notice.Level,
		"publishTime": notice.PublishTime,
	}

	if notice.TargetType == 1 {
		onlineUsers := sseService.GetOnlineUsers()
		for _, u := range onlineUsers {
			sseService.SendToUser(u.Username, "notice", noticeData)
		}
	} else if len(targetUsers) > 0 {
		// TODO: 将用户ID转换为用户名后推送
	}
}

// PublishNotice 发布通知
func PublishNotice(id int64, publisherID int64) error {
	notice, err := repository.GetNoticeByID(id)
	if err != nil {
		return errs.NotFound("通知不存在")
	}

	now := time.Now()
	if err := repository.UpdateNoticeFields(int64(notice.ID), map[string]interface{}{
		"publish_status": 1,
		"publisher_id":   publisherID,
		"publish_time":   now,
		"revoke_time":    nil,
	}); err != nil {
		return errs.SystemError("发布通知失败")
	}

	// 发布时先删除该通知之前的用户通知数据（重新发布场景）
	database.DB.Exec("DELETE FROM sys_user_notice WHERE notice_id = ?", id)

	notice.Status = 1
	notice.PublisherID = types.BigInt(publisherID)
	pt := types.LocalTime(now)
	notice.PublishTime = &pt

	// 推送通知
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

	now := time.Now()
	if err := repository.UpdateNoticeFields(int64(notice.ID), map[string]interface{}{
		"publish_status": -1,
		"revoke_time":   now,
	}); err != nil {
		return errs.SystemError("撤回通知失败")
	}

	// 撤回时删除用户通知状态记录
	database.DB.Exec("DELETE FROM sys_user_notice WHERE notice_id = ?", id)

	// 通知前端移除该通知
	sseService := message.GetSseService()
	if sseService != nil {
		onlineUsers := sseService.GetOnlineUsers()
		for _, u := range onlineUsers {
			sseService.SendToUser(u.Username, "notice-revoke", map[string]interface{}{"id": id})
		}
	}

	return nil
}
