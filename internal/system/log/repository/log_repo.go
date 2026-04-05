package repository

import (
	"strings"
	"time"

	"youlai-gin/internal/system/log/model"
	pkgDatabase "youlai-gin/internal/common/database"
	"youlai-gin/internal/common/database"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/types"
)

// GetLogPage 获取日志分页列表
func GetLogPage(query *model.LogQuery) ([]model.LogPageVO, int64, error) {
	var logs []struct {
		ID            int64     `gorm:"column:id"`
		Module        int       `gorm:"column:module"`
		ActionType    int       `gorm:"column:action_type"`
		Title         string    `gorm:"column:title"`
		Content       string    `gorm:"column:content"`
		OperatorID    int64     `gorm:"column:operator_id"`
		OperatorName  string    `gorm:"column:operator_name"`
		Status        int       `gorm:"column:status"`
		RequestURI    string    `gorm:"column:request_uri"`
		RequestMethod string    `gorm:"column:request_method"`
		IP            string    `gorm:"column:ip"`
		Region        string    `gorm:"column:region"`
		Device        string    `gorm:"column:device"`
		Browser       string    `gorm:"column:browser"`
		OS            string    `gorm:"column:os"`
		ExecutionTime int       `gorm:"column:execution_time"`
		ErrorMsg      string    `gorm:"column:error_msg"`
		CreateTime    time.Time `gorm:"column:create_time"`
	}
	var total int64

	db := database.DB.Table("sys_log t1").
		Select("t1.id, t1.module, t1.action_type, t1.title, t1.content, " +
			"t1.operator_id, t1.operator_name, t1.status, t1.request_uri, t1.request_method, t1.ip, " +
			"CONCAT(t1.province,' ', t1.city) as region, t1.device, t1.browser, t1.os, " +
			"t1.execution_time, t1.error_msg, t1.create_time")

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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("t1.create_time DESC").
		Find(&logs).Error

	// 转换为 VO 并处理枚举标签
	result := make([]model.LogPageVO, len(logs))
	for i, log := range logs {
		moduleLabel := enums.LogModuleDesc[enums.LogModule(log.Module)]
		if moduleLabel == "" {
			moduleLabel = "其他"
		}
		actionTypeLabel := enums.ActionTypeDesc[enums.ActionType(log.ActionType)]
		if actionTypeLabel == "" {
			actionTypeLabel = "其他"
		}
		result[i] = model.LogPageVO{
			ID:            types.BigInt(log.ID),
			Module:        moduleLabel,
			ActionType:    actionTypeLabel,
			Title:         log.Title,
			Content:       log.Content,
			OperatorID:    types.BigInt(log.OperatorID),
			OperatorName:  log.OperatorName,
			Status:        log.Status,
			RequestURI:    log.RequestURI,
			RequestMethod: log.RequestMethod,
			IP:            log.IP,
			Region:        log.Region,
			Device:        log.Device,
			Browser:       log.Browser,
			OS:            log.OS,
			ExecutionTime: log.ExecutionTime,
			ErrorMsg:      log.ErrorMsg,
			CreateTime:    types.LocalTime(log.CreateTime),
		}
	}

	return result, total, err
}

// GetVisitTrend 获取访问趋势
func GetVisitTrend(startDate, endDate time.Time) (*model.VisitTrendVO, error) {
	// 生成日期列表
	dates := make([]string, 0)
	pvList := make([]int64, 0)
	uvList := make([]int64, 0)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dates = append(dates, dateStr)

		// PV（总访问量）
		var pv int64
		database.DB.Table("sys_log").
			Where("DATE(create_time) = ?", dateStr).
			Count(&pv)
		pvList = append(pvList, pv)

		// UV（独立访客）
		var uvCount int64
		database.DB.Table("sys_log").
			Where("DATE(create_time) = ?", dateStr).
			Distinct("ip").
			Count(&uvCount)
		uvList = append(uvList, uvCount)
	}

	return &model.VisitTrendVO{
		Dates:  dates,
		PvList: pvList,
		UvList: uvList,
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
		Count(&stats.TodayPvCount)

	database.DB.Table("sys_log").
		Where("DATE(create_time) = ?", today).
		Distinct("operator_id").
		Count(&stats.TodayUvCount)

	// 总统计
	database.DB.Table("sys_log").Count(&stats.TotalPvCount)
	database.DB.Table("sys_log").Distinct("operator_id").Count(&stats.TotalUvCount)

	_ = weekStart
	_ = monthStart
	stats.UvGrowthRate = 0
	stats.PvGrowthRate = 0

	return stats, nil
}
