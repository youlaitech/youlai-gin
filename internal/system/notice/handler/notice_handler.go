package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/notice/model"
	"youlai-gin/internal/system/notice/service"
	"youlai-gin/pkg/types"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册通知公告路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/notices", GetNoticePage)
	r.POST("/notices", SaveNotice)
	r.GET("/notices/:id/form", GetNoticeForm)
	r.GET("/notices/:id/detail", GetNoticeDetail)
	r.PUT("/notices/:id", UpdateNotice)
	r.PUT("/notices/:id/publish", PublishNotice)
	r.PUT("/notices/:id/revoke", RevokeNotice)
	r.DELETE("/notices/:ids", DeleteNotices)
	r.GET("/notices/my", GetMyNoticePage)
	r.PUT("/notices/read-all", ReadAllNotices)
	r.GET("/notices/unread-count", GetUnreadCount)
}

// GetNoticePage 通知公告分页列表
// @Summary 通知公告分页
// @Tags 09.通知公告
// @Router /api/v1/notices [get]
func GetNoticePage(c *gin.Context) {
	var query model.NoticeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetNoticePage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// SaveNotice 新增通知公告
// @Summary 新增通知公告
// @Tags 09.通知公告
// @Router /api/v1/notices [post]
func SaveNotice(c *gin.Context) {
	var form model.NoticeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.SaveNotice(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetNoticeForm 获取通知公告表单数据
// @Summary 通知公告表单
// @Tags 09.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/form [get]
func GetNoticeForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	notice, err := service.GetNoticeByID(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, notice)
}

// GetNoticeDetail 阅读获取通知公告详情
// @Summary 通知公告详情
// @Tags 09.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/detail [get]
func GetNoticeDetail(c *gin.Context) {
	idStr := c.Param("id")
	noticeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	// 获取当前用户ID
	userID, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 获取通知详情
	notice, err := service.GetNoticeByID(noticeID)
	if err != nil {
		c.Error(err)
		return
	}

	// 标记为已读
	go service.MarkNoticeAsRead(noticeID, userID)

	response.Ok(c, notice)
}

// UpdateNotice 修改通知公告
// @Summary 修改通知公告
// @Tags 09.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id} [put]
func UpdateNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	var form model.NoticeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	form.ID = types.BigInt(id)
	if err := service.SaveNotice(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// PublishNotice 发布通知公告
// @Summary 发布通知公告
// @Tags 09.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/publish [put]
func PublishNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	userID, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := service.PublishNotice(id, userID); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "发布成功")
}

// RevokeNotice 撤回通知公告
// @Summary 撤回通知公告
// @Tags 09.通知公告
// @Param id path int true "公告ID"
// @Router /api/v1/notices/{id}/revoke [put]
func RevokeNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	if err := service.RevokeNotice(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "撤回成功")
}

// DeleteNotices 删除通知公告（支持批量）
// @Summary 删除通知公告
// @Tags 09.通知公告
// @Param ids path string true "公告ID列表"
// @Router /api/v1/notices/{ids} [delete]
func DeleteNotices(c *gin.Context) {
	idsStr := c.Param("ids")
	if idsStr == "" {
		response.Fail(c, "ID不能为空")
		return
	}

	// 支持批量删除，ids格式：1,2,3
	idStrArr := strings.Split(idsStr, ",")
	for _, idStr := range idStrArr {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			response.Fail(c, "ID格式错误")
			return
		}
		if err := service.DeleteNotice(id); err != nil {
			c.Error(err)
			return
		}
	}

	response.OkMsg(c, "删除成功")
}

// GetMyNoticePage 获取我的通知公告分页列表
// @Summary 我的通知公告
// @Tags 09.通知公告
// @Router /api/v1/notices/my [get]
func GetMyNoticePage(c *gin.Context) {
	// 获取当前用户ID
	userID, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var query model.UserNoticeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetUserNoticePage(userID, &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// ReadAllNotices 全部已读
// @Summary 通知全部已读
// @Tags 09.通知公告
// @Router /api/v1/notices/read-all [put]
func ReadAllNotices(c *gin.Context) {
	// 获取当前用户ID
	userID, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 这里需要在 repository 层添加全部已读的方法
	// 暂时返回成功，后续补充实现
	_ = userID
	response.OkMsg(c, "全部已读成功")
}

// GetUnreadCount 获取未读通知数量
// @Summary 未读通知数量
// @Tags 09.通知公告
// @Router /api/v1/notices/unread-count [get]
func GetUnreadCount(c *gin.Context) {
	// 获取当前用户ID
	userID, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	count, err := service.GetUnreadCount(userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, map[string]interface{}{
		"count": count,
	})
}
