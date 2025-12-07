package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册日志路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/logs/page", GetLogPage)
	r.GET("/logs/visit-trend", GetVisitTrend)
	r.GET("/logs/visit-stats", GetVisitStats)
}

// GetLogPage 日志分页列表
func GetLogPage(c *gin.Context) {
	var query model.LogPageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetLogPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// GetVisitTrend 获取访问趋势
func GetVisitTrend(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		response.Fail(c, "开始时间和结束时间不能为空")
		return
	}

	result, err := service.GetVisitTrend(startDate, endDate)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// GetVisitStats 获取访问统计
func GetVisitStats(c *gin.Context) {
	result, err := service.GetVisitStats()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
