package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	sse "youlai-gin/internal/message/service"
	"youlai-gin/internal/system/notice/model"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/gormx"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// Repository 通知公告数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.NoticeQuery) ([]model.Notice, int64, error)
	Get(ctx context.Context, id int64) (*model.Notice, error)
	Create(ctx context.Context, notice *model.Notice) error
	Update(ctx context.Context, id int64, patch map[string]any) error
	Delete(ctx context.Context, id int64) error
	DeleteUserNoticesByNoticeID(ctx context.Context, noticeID int64) error
	UserNoticePage(ctx context.Context, userID int64, query *model.UserNoticeQuery) ([]model.Notice, int64, error)
	MarkRead(ctx context.Context, noticeID, userID int64) error
	UnreadCount(ctx context.Context, userID int64) (int64, error)
}

// Service 通知公告业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Page 通知分页列表
func (s *Service) Page(ctx context.Context, query *model.NoticeQuery) (*baseModel.PagedData, error) {
	list, total, err := s.repo.Page(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询通知列表失败")
	}

	return &baseModel.PagedData{List: list, Total: total}, nil
}

// Get 根据 ID 获取通知
func (s *Service) Get(ctx context.Context, id int64) (*model.Notice, error) {
	return s.repo.Get(ctx, id)
}

// Create 新增通知
// 发布状态（status=1）未指定发布时间时默认当前时间；目标用户切片转 JSON 存储。
func (s *Service) Create(ctx context.Context, form *model.NoticeForm) error {
	pt, err := parsePublishTime(form.PublishTime)
	if err != nil {
		return errs.BadRequest("发布时间格式错误")
	}

	notice := &model.Notice{
		Title:      form.Title,
		Content:    form.Content,
		Type:       form.Type,
		Level:      form.Level,
		Status:     form.Status,
		TargetType: form.TargetType,
	}
	if pt != nil {
		notice.PublishTime = pt
	} else if form.Status == 1 {
		now := types.Now()
		notice.PublishTime = &now
	}
	if len(form.TargetUsers) > 0 {
		if b, err := json.Marshal(form.TargetUsers); err == nil {
			notice.TargetUsers = string(b)
		}
	}

	if err := s.repo.Create(ctx, notice); err != nil {
		return errs.SystemError("创建通知失败")
	}

	if form.Status == 1 {
		pushed := &model.Notice{ID: form.ID, Title: form.Title, Type: form.Type, Level: form.Level, Status: form.Status}
		go pushNotice(pushed, form.TargetUsers)
	}

	return nil
}

// Update 更新通知
// 由 BuildPatchMap(form) 生成部分更新映射，并补充 publish_time、target_user_ids 等需要转换的字段。
func (s *Service) Update(ctx context.Context, id int64, form *model.NoticeForm) error {
	patch := gormx.BuildPatchMap(form)

	pt, err := parsePublishTime(form.PublishTime)
	if err != nil {
		return errs.BadRequest("发布时间格式错误")
	}
	if pt != nil {
		patch["publish_time"] = pt
	} else if v, ok := patch["publish_status"].(int); ok && v == 1 {
		now := types.Now()
		patch["publish_time"] = &now
	}

	if len(form.TargetUsers) > 0 {
		if b, err := json.Marshal(form.TargetUsers); err == nil {
			patch["target_user_ids"] = string(b)
		}
	}

	if err := s.repo.Update(ctx, id, patch); err != nil {
		return errs.SystemError("更新通知失败")
	}

	if form.Status == 1 {
		pushed := &model.Notice{ID: types.BigInt(id), Title: form.Title, Type: form.Type, Level: form.Level, Status: form.Status}
		go pushNotice(pushed, form.TargetUsers)
	}

	return nil
}

// Delete 删除通知（支持批量）
func (s *Service) Delete(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		if err := s.repo.Delete(ctx, id); err != nil {
			return errs.SystemError("删除通知失败")
		}
	}
	return nil
}

// Publish 发布通知
func (s *Service) Publish(ctx context.Context, id int64, publisherID int64) error {
	notice, err := s.repo.Get(ctx, id)
	if err != nil {
		return errs.NotFound("通知不存在")
	}

	now := time.Now()
	if err := s.repo.Update(ctx, int64(notice.ID), map[string]any{
		"publish_status": 1,
		"publisher_id":   publisherID,
		"publish_time":   now,
		"revoke_time":    nil,
	}); err != nil {
		return errs.SystemError("发布通知失败")
	}

	// 发布时先删除该通知之前的用户通知数据（重新发布场景）
	_ = s.repo.DeleteUserNoticesByNoticeID(ctx, id)

	notice.Status = 1
	notice.PublisherID = types.BigInt(publisherID)
	pt := types.LocalTime(now)
	notice.PublishTime = &pt

	var targetUsers []types.BigInt
	if notice.TargetUsers != "" {
		json.Unmarshal([]byte(notice.TargetUsers), &targetUsers)
	}
	go pushNotice(notice, targetUsers)

	return nil
}

// Revoke 撤回通知
func (s *Service) Revoke(ctx context.Context, id int64) error {
	notice, err := s.repo.Get(ctx, id)
	if err != nil {
		return errs.NotFound("通知不存在")
	}

	if notice.Status != 1 {
		return errs.BadRequest("通知未发布或已撤回")
	}

	if err := s.repo.Update(ctx, int64(notice.ID), map[string]any{
		"publish_status": -1,
		"revoke_time":    time.Now(),
	}); err != nil {
		return errs.SystemError("撤回通知失败")
	}

	// 撤回时删除用户通知状态记录
	_ = s.repo.DeleteUserNoticesByNoticeID(ctx, id)

	// 通知前端移除该通知
	sseService := sse.GetSseService()
	if sseService != nil {
		for _, u := range sseService.GetOnlineUsers() {
			sseService.SendToUser(u.Username, "notice-revoke", map[string]interface{}{"id": id})
		}
	}

	return nil
}

// UserPage 用户通知分页列表
func (s *Service) UserPage(ctx context.Context, userID int64, query *model.UserNoticeQuery) (*baseModel.PagedData, error) {
	list, total, err := s.repo.UserNoticePage(ctx, userID, query)
	if err != nil {
		return nil, errs.SystemError("查询用户通知列表失败")
	}

	return &baseModel.PagedData{List: list, Total: total}, nil
}

// MarkRead 标记通知为已读
func (s *Service) MarkRead(ctx context.Context, noticeID, userID int64) error {
	if err := s.repo.MarkRead(ctx, noticeID, userID); err != nil {
		return errs.SystemError("标记通知已读失败")
	}
	return nil
}

// UnreadCount 用户未读通知数量
func (s *Service) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	return s.repo.UnreadCount(ctx, userID)
}

// parsePublishTime 解析表单发布时间字符串，空串返回 nil
func parsePublishTime(s string) (*types.LocalTime, error) {
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

// pushNotice 推送通知（SSE）
func pushNotice(notice *model.Notice, targetUsers []types.BigInt) {
	sseService := sse.GetSseService()
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
		for _, u := range sseService.GetOnlineUsers() {
			sseService.SendToUser(u.Username, "notice", noticeData)
		}
	} else if len(targetUsers) > 0 {
		// TODO: 将用户ID转换为用户名后推送
	}
}
