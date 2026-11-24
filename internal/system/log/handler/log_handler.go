package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/errs"
)

// Handler 日志管理 HTTP 处理器
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册日志路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/logs", auth.RequirePermission("sys:log:list"), h.Page)
	r.GET("/analytics/visits/summary", h.VisitSummary)
	r.GET("/analytics/visits/trend", h.VisitTrend)
	r.GET("/analytics/users/summary", h.UserSummary)
}

// Page 日志分页列表
// @Summary 日志分页
// @Tags 09.日志接口
// @Router /api/v1/logs [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.LogQuery
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

// VisitSummary 访问汇总
// @Summary 访问汇总
// @Tags 15.统计接口
// @Router /api/v1/analytics/visits/summary [get]
func (h *Handler) VisitSummary(c *gin.Context) {
	result, err := h.svc.VisitSummary(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

// VisitTrend 访问趋势
// @Summary 访问趋势
// @Tags 15.统计接口
// @Router /api/v1/analytics/visits/trend [get]
func (h *Handler) VisitTrend(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if startDate == "" || endDate == "" {
		c.Error(errs.BadRequest("开始日期和结束日期不能为空"))
		return
	}
	result, err := h.svc.VisitTrend(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

// UserSummary 用户汇总
// @Summary 用户汇总
// @Tags 15.统计接口
// @Router /api/v1/analytics/users/summary [get]
func (h *Handler) UserSummary(c *gin.Context) {
	result, err := h.svc.UserSummary(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}
