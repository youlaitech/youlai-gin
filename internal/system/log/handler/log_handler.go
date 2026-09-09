package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
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
	r.GET("/logs", h.Page)
	r.GET("/logs/analytics/trend", h.VisitTrend)
	r.GET("/logs/analytics/overview", h.VisitOverview)
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

// VisitTrend 访问趋势统计
// @Summary 访问趋势
// @Tags 09.日志接口
// @Router /api/v1/logs/analytics/trend [get]
func (h *Handler) VisitTrend(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		c.Error(errs.BadRequest("开始时间和结束时间不能为空"))
		return
	}

	result, err := h.svc.VisitTrend(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// VisitOverview 访问统计概览
// @Summary 访问统计概览
// @Tags 09.日志接口
// @Router /api/v1/logs/analytics/overview [get]
func (h *Handler) VisitOverview(c *gin.Context) {
	result, err := h.svc.VisitStats(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
