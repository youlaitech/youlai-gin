package service

import (
	"context"
	"time"

	"youlai-gin/internal/system/log/model"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
)

// Repository 日志数据访问接口（依赖倒置：Service 定义接口）
type Repository interface {
	Page(ctx context.Context, query *model.LogQuery) ([]model.LogPageVO, int64, error)
	VisitTrend(ctx context.Context, startDate, endDate time.Time) (*model.VisitTrendVO, error)
	VisitStats(ctx context.Context) (*model.VisitStatsVO, error)
}

// Service 日志业务逻辑层
type Service struct {
	repo Repository
}

// NewService 创建 Service 实例
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Page 日志分页列表
func (s *Service) Page(ctx context.Context, query *model.LogQuery) (*baseModel.PagedData, error) {
	logs, total, err := s.repo.Page(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询日志列表失败")
	}

	return &baseModel.PagedData{List: logs, Total: total}, nil
}

// VisitTrend 访问趋势（校验日期范围后按日统计）
func (s *Service) VisitTrend(ctx context.Context, startDate, endDate string) (*model.VisitTrendVO, error) {
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

	if end.Sub(start).Hours() > 90*24 {
		return nil, errs.BadRequest("查询范围不能超过90天")
	}

	return s.repo.VisitTrend(ctx, start, end)
}

// VisitStats 访问统计概览
func (s *Service) VisitStats(ctx context.Context) (*model.VisitStatsVO, error) {
	return s.repo.VisitStats(ctx)
}
