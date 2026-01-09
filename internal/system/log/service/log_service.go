package service

import (
	"time"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/internal/system/log/repository"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
)

// GetLogPage 获取日志分页列表
func GetLogPage(query *model.LogQuery) (*common.PagedData, error) {
	logs, total, err := repository.GetLogPage(query)
	if err != nil {
		return nil, errs.SystemError("查询日志列表失败")
	}

	pageMeta := common.NewPageMeta(query.PageNum, query.PageSize, total)
	return &common.PagedData{Data: logs, Page: pageMeta}, nil
}

// GetVisitTrend 获取访问趋势
func GetVisitTrend(startDate, endDate string) (*model.VisitTrendVO, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errs.BadRequest("开始日期格式错误")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errs.BadRequest("结束日期格式错误")
	}

	if start.After(end) {
		return nil, errs.BadRequest("开始日期不能晚于结束日期")
	}

	// 限制查询范围不超过90天
	if end.Sub(start).Hours() > 90*24 {
		return nil, errs.BadRequest("查询范围不能超过90天")
	}

	return repository.GetVisitTrend(start, end)
}

// GetVisitStats 获取访问统计
func GetVisitStats() (*model.VisitStatsVO, error) {
	return repository.GetVisitStats()
}
