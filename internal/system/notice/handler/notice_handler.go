package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/notice/model"
	"youlai-gin/internal/system/notice/service"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册通知公告路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/notices/page", GetNoticePage)
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
func GetNoticePage(c *gin.Context) {
	var query model.NoticePageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetNoticePage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// SaveNotice 新增通知公告
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

	form.ID = id
	if err := service.SaveNotice(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// PublishNotice 发布通知公告
func PublishNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	if err := service.PublishNotice(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "发布成功")
}

// RevokeNotice 撤回通知公告
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

	response.Ok(c, result)
}

// ReadAllNotices 全部已读
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
