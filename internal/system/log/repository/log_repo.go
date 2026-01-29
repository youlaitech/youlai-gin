package repository

import (
	"strings"
	"time"

	"youlai-gin/pkg/database"
	"youlai-gin/internal/system/log/model"
	pkgDatabase "youlai-gin/pkg/database"
)

// GetLogPage 获取日志分页列表
func GetLogPage(query *model.LogQuery) ([]model.LogPageVO, int64, error) {
	var logs []model.LogPageVO
	var total int64

	db := database.DB.Table("sys_log t1").
		Select("t1.id, t1.module, t1.content, t1.request_uri, t1.request_method as method, t1.ip, " +
			"CONCAT(t1.province,' ', t1.city) as region, t1.execution_time, " +
			"CONCAT(t1.browser,' ', t1.browser_version) as browser, t1.os, t1.create_by, t1.create_time, " +
			"t2.nickname as operator").
		Joins("LEFT JOIN sys_user t2 ON t1.create_by = t2.id")

	// 查询条件
	if query.Keywords != "" {
		keyword := "%" + query.Keywords + "%"
		db = db.Where("(t1.content LIKE ? OR t1.ip LIKE ? OR t2.nickname LIKE ?)", keyword, keyword, keyword)
	}

	if len(query.CreateTime) == 2 {
		startTime := strings.TrimSpace(query.CreateTime[0])
		endTime := strings.TrimSpace(query.CreateTime[1])
		if startTime != "" {
			if len(startTime) == 10 {
				startTime += " 00:00:00"
			}
			db = db.Where("t1.create_time >= ?", startTime)
		}
		if endTime != "" {
			if len(endTime) == 10 {
				endTime += " 23:59:59"
			}
			db = db.Where("t1.create_time <= ?", endTime)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("t1.create_time DESC").
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
