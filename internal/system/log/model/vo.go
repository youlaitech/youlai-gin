package model

import "youlai-gin/pkg/types"

// LogPageVO 日志分页VO
type LogPageVO struct {
	ID         types.BigInt `json:"id"`
	Module     string       `json:"module"`
	Operation  string       `json:"operation"`
	Method     string       `json:"method"`
	Path       string       `json:"path"`
	UserID     types.BigInt `json:"userId"`
	Username   string       `json:"username"`
	IP         string       `json:"ip"`
	Status     int          `json:"status"`
	Duration   int64        `json:"duration"`
	ErrorMsg   string       `json:"errorMsg"`
	CreateTime string       `json:"createTime"`
}

// VisitTrendVO 访问趋势VO
type VisitTrendVO struct {
	Dates []string `json:"dates"` // 日期列表
	PVs   []int64  `json:"pvs"`   // 访问量列表
	UVs   []int64  `json:"uvs"`   // 独立访客列表
	IPs   []int64  `json:"ips"`   // 独立IP列表
}

// VisitStatsVO 访问统计VO
type VisitStatsVO struct {
	TodayPV int64 `json:"todayPv"` // 今日访问量
	TodayUV int64 `json:"todayUv"` // 今日独立访客
	TodayIP int64 `json:"todayIp"` // 今日独立IP
	WeekPV  int64 `json:"weekPv"`  // 本周访问量
	WeekUV  int64 `json:"weekUv"`  // 本周独立访客
	MonthPV int64 `json:"monthPv"` // 本月访问量
	MonthUV int64 `json:"monthUv"` // 本月独立访客
	TotalPV int64 `json:"totalPv"` // 总访问量
	TotalUV int64 `json:"totalUv"` // 总独立访客
}
