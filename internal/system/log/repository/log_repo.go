package repository

import (
	"time"

	"youlai-gin/internal/database"
	"youlai-gin/internal/system/log/model"
	pkgDatabase "youlai-gin/pkg/database"
)

// GetLogPage 获取日志分页列表
func GetLogPage(query *model.LogPageQuery) ([]model.LogPageVO, int64, error) {
	var logs []model.LogPageVO
	var total int64

	db := database.DB.Table("sys_log")

	// 查询条件
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}

	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if query.StartTime != "" {
		db = db.Where("create_time >= ?", query.StartTime)
	}

	if query.EndTime != "" {
		db = db.Where("create_time <= ?", query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("create_time DESC").
		Find(&logs).Error

	return logs, total, err
}

// GetVisitTrend 获取访问趋势
func GetVisitTrend(startDate, endDate time.Time) (*model.VisitTrendVO, error) {
	// 生成日期列表
	dates := make([]string, 0)
	pvs := make([]int64, 0)
	uvs := make([]int64, 0)
	ips := make([]int64, 0)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dates = append(dates, dateStr)

		// PV（总访问量）
		var pv int64
		database.DB.Table("sys_log").
			Where("DATE(create_time) = ?", dateStr).
			Count(&pv)
		pvs = append(pvs, pv)

		// UV（独立用户）
		var uv int64
		database.DB.Table("sys_log").
			Where("DATE(create_time) = ?", dateStr).
			Distinct("create_by").
			Count(&uv)
		uvs = append(uvs, uv)

		// IP（独立IP）
		var ipCount int64
		database.DB.Table("sys_log").
			Where("DATE(create_time) = ?", dateStr).
			Distinct("ip").
			Count(&ipCount)
		ips = append(ips, ipCount)
	}

	return &model.VisitTrendVO{
		Dates: dates,
		PVs:   pvs,
		UVs:   uvs,
		IPs:   ips,
	}, nil
}

// GetVisitStats 获取访问统计
func GetVisitStats() (*model.VisitStatsVO, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	stats := &model.VisitStatsVO{}

	// 今日统计
	database.DB.Table("sys_log").
		Where("DATE(create_time) = ?", today).
		Count(&stats.TodayPV)

	database.DB.Table("sys_log").
		Where("DATE(create_time) = ?", today).
		Distinct("create_by").
		Count(&stats.TodayUV)

	database.DB.Table("sys_log").
		Where("DATE(create_time) = ?", today).
		Distinct("ip").
		Count(&stats.TodayIP)

	// 本周统计
	database.DB.Table("sys_log").
		Where("DATE(create_time) >= ?", weekStart).
		Count(&stats.WeekPV)

	database.DB.Table("sys_log").
		Where("DATE(create_time) >= ?", weekStart).
		Distinct("create_by").
		Count(&stats.WeekUV)

	// 本月统计
	database.DB.Table("sys_log").
		Where("DATE(create_time) >= ?", monthStart).
		Count(&stats.MonthPV)

	database.DB.Table("sys_log").
		Where("DATE(create_time) >= ?", monthStart).
		Distinct("create_by").
		Count(&stats.MonthUV)

	// 总统计
	database.DB.Table("sys_log").Count(&stats.TotalPV)
	database.DB.Table("sys_log").Distinct("create_by").Count(&stats.TotalUV)

	return stats, nil
}
