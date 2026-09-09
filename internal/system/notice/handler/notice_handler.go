package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/system/notice/model"
	"youlai-gin/internal/system/notice/service"
	"youlai-gin/pkg/types"
)

// Handler 通知公告 HTTP 处理器
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册通知公告路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/notices", h.Page)
	r.POST("/notices", h.Create)
	r.GET("/notices/:id/form", h.GetForm)
	r.GET("/notices/:id/detail", h.GetDetail)
	r.PUT("/notices/:id", h.Update)
	r.PUT("/notices/:id/publish", h.Publish)
	r.PUT("/notices/:id/revoke", h.Revoke)
	r.DELETE("/notices/:ids", h.Delete)
	r.GET("/notices/my", h.MyPage)
	r.PUT("/notices/read-all", h.ReadAll)
	r.GET("/notices/unread-count", h.UnreadCount)
}

// Page 通知公告分页列表
// @Summary 通知公告分页
// @Tags 08.通知公告
// @Router /api/v1/notices [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.NoticeQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	result, err := h.svc.Page(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// Create 新增通知公告
// @Summary 新增通知公告
// @Tags 08.通知公告
// @Router /api/v1/notices [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.NoticeForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Create(appContext.OperatorCtx(c), &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetForm 获取通知公告表单数据
// @Summary 通知公告表单
// @Tags 08.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "通知")
	if err != nil {
		c.Error(err)
		return
	}

	notice, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, notice)
}

// GetDetail 阅读获取通知公告详情（自动标记已读）
// @Summary 通知公告详情
// @Tags 08.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/detail [get]
func (h *Handler) GetDetail(c *gin.Context) {
	noticeID, err := appContext.ParsePathParam(c, "id", "通知")
	if err != nil {
		c.Error(err)
		return
	}

	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	notice, err := h.svc.Get(c.Request.Context(), noticeID)
	if err != nil {
		c.Error(err)
		return
	}

	// 标记已读独立于请求上下文，避免异步执行时请求上下文已取消
	go h.svc.MarkRead(context.Background(), noticeID, userID)

	response.Ok(c, notice)
}

// Update 修改通知公告
// @Summary 修改通知公告
// @Tags 08.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "通知")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.NoticeForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = types.BigInt(id)
	if err := h.svc.Update(appContext.OperatorCtx(c), id, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// Publish 发布通知公告
// @Summary 发布通知公告
// @Tags 08.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/publish [put]
func (h *Handler) Publish(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "通知")
	if err != nil {
		c.Error(err)
		return
	}

	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Publish(appContext.OperatorCtx(c), id, userID); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "发布成功")
}

// Revoke 撤回通知公告
// @Summary 撤回通知公告
// @Tags 08.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/revoke [put]
func (h *Handler) Revoke(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "通知")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Revoke(appContext.OperatorCtx(c), id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "撤回成功")
}

// Delete 删除通知公告（支持批量）
// @Summary 删除通知公告
// @Tags 08.通知公告
// @Param ids path string true "公告ID列表"
// @Router /api/v1/notices/{ids} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ids, err := appContext.ParseIntList(c.Param("ids"), "通知")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Delete(appContext.OperatorCtx(c), ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// MyPage 获取我的通知公告分页列表
// @Summary 我的通知公告
// @Tags 08.通知公告
// @Router /api/v1/notices/my [get]
func (h *Handler) MyPage(c *gin.Context) {
	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var query model.UserNoticeQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	result, err := h.svc.UserPage(c.Request.Context(), userID, &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// ReadAll 全部已读
// @Summary 通知全部已读
// @Tags 08.通知公告
// @Router /api/v1/notices/read-all [put]
func (h *Handler) ReadAll(c *gin.Context) {
	response.OkMsg(c, "全部已读成功")
}

// UnreadCount 获取未读通知数量
// @Summary 未读通知数量
// @Tags 08.通知公告
// @Router /api/v1/notices/unread-count [get]
func (h *Handler) UnreadCount(c *gin.Context) {
	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, map[string]interface{}{
		"count": count,
	})
}
