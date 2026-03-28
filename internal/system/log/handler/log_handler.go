package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

// RegisterRoutes 注册日志路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/logs", GetLogPage)
	r.GET("/logs/visits/trend", GetVisitTrend)
	r.GET("/logs/visits/overview", GetVisitOverview)
}

// GetLogPage 日志分页列表
// @Summary 日志分页
// @Tags 09.日志接口
// @Router /api/v1/logs [get]
func GetLogPage(c *gin.Context) {
	var query model.LogQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	result, err := service.GetLogPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetVisitTrend 访问趋势统计
// @Summary 访问趋势
// @Tags 09.日志接口
// @Router /api/v1/logs/visits/trend [get]
func GetVisitTrend(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		c.Error(errs.BadRequest("开始时间和结束时间不能为空"))
		return
	}

	result, err := service.GetVisitTrend(startDate, endDate)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// GetVisitOverview 访问统计概览
// @Summary 访问统计概览
// @Tags 09.日志接口
// @Router /api/v1/logs/views [get]
func GetVisitOverview(c *gin.Context) {
	result, err := service.GetVisitStats()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
