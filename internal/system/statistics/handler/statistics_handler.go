package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册统计分析路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("statistics/visits/trend", GetVisitTrend)
	r.GET("statistics/visits/overview", GetVisitOverview)
}

// GetVisitTrend 访问趋势统计
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

// GetVisitOverview 访问概览统计
func GetVisitOverview(c *gin.Context) {
	result, err := service.GetVisitStats()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
