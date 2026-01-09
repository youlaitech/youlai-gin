package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册日志路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/logs", GetLogPage)
}

// GetLogPage 日志分页列表
func GetLogPage(c *gin.Context) {
	var query model.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetLogPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}
