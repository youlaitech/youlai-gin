package model

import "youlai-gin/pkg/types"

// LogPageVO 日志分页VO
type LogPageVO struct {
	ID            types.BigInt    `json:"id"`
	Module        string          `json:"module"`
	ActionType    string          `json:"actionType"`
	Title         string          `json:"title"`
	Content       string          `json:"content"`
	OperatorID    types.BigInt    `json:"operatorId"`
	OperatorName  string          `json:"operatorName"`
	Status        int             `json:"status"`
	RequestURI    string          `json:"requestUri"`
	RequestMethod string          `json:"requestMethod"`
	IP            string          `json:"ip"`
	Region        string          `json:"region"`
	Device        string          `json:"device"`
	Browser       string          `json:"browser"`
	OS            string          `json:"os"`
	ExecutionTime int             `json:"executionTime"`
	ErrorMsg      string          `json:"errorMsg"`
	CreateTime    types.LocalTime `json:"createTime"`
}

// VisitTrendVO 访问趋势VO
type VisitTrendVO struct {
	Dates  []string `json:"dates"`  // 日期列表
	PvList []int64  `json:"pvList"` // 浏览量(PV)
	UvList []int64  `json:"uvList"` // 访客数(UV)
}

// VisitStatsVO 访问统计VO（对齐前端 vue3-element-admin 字段）
type VisitStatsVO struct {
	TodayUvCount int64 `json:"todayUvCount"` // 今日独立访客数 (UV)
	TotalUvCount int64 `json:"totalUvCount"` // 累计独立访客数 (UV)
	UvGrowthRate int64 `json:"uvGrowthRate"` // 独立访客增长率（暂未计算，返回 0）
	TodayPvCount int64 `json:"todayPvCount"` // 今日页面浏览量 (PV)
	TotalPvCount int64 `json:"totalPvCount"` // 累计页面浏览量 (PV)
	PvGrowthRate int64 `json:"pvGrowthRate"` // 页面浏览量增长率（暂未计算，返回 0）
}
