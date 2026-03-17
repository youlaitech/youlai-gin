package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/service"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

// RegisterRoutes 注册日志路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/logs", GetLogPage)
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
