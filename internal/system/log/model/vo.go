package model

import "youlai-gin/pkg/types"

// LogPageVO 日志分页VO
type LogPageVO struct {
	ID            types.BigInt    `json:"id"` // 主键
	Module        string          `json:"module"`
	ActionType    string          `json:"actionType"`
	Title         string          `json:"title"`
	Content       string          `json:"content"`
	OperatorID    types.BigInt    `json:"operatorId"`
	OperatorName  string          `json:"operatorName"`
	Status        int             `json:"status"` // 状态(1启用0禁用)
	RequestURI    string          `json:"requestUri"`
	RequestMethod string          `json:"requestMethod"`
	IP            string          `json:"ip"`
	Region        string          `json:"region"`
	Device        string          `json:"device"`
	Browser       string          `json:"browser"`
	OS            string          `json:"os"`
	ExecutionTime int             `json:"executionTime"`
	ErrorMsg      string          `json:"errorMsg"`
	CreateTime    types.LocalTime `json:"createTime"` // 创建时间
}

// VisitTrendPointVO 单日访问量
type VisitTrendPointVO struct {
	Date string `json:"date"` // 日期 yyyy-MM-dd
	Pv   int64  `json:"pv"`   // 浏览量
	Uv   int64  `json:"uv"`   // 访客数
}

// VisitTrendVO 访问趋势VO
type VisitTrendVO struct {
	StartDate string               `json:"startDate"` // 查询开始日期
	EndDate   string               `json:"endDate"`   // 查询结束日期
	Points    []VisitTrendPointVO  `json:"points"`    // 逐日访问量
}

// VisitSummaryVO 访问汇总VO
type VisitSummaryVO struct {
	TodayUv      int64    `json:"todayUv"`      // 今日访客数
	YesterdayUv  int64    `json:"yesterdayUv"`  // 昨日访客数
	TotalUv      int64    `json:"totalUv"`      // 累计访客数
	TodayPv      int64    `json:"todayPv"`      // 今日浏览量
	YesterdayPv  int64    `json:"yesterdayPv"`  // 昨日浏览量
	TotalPv      int64    `json:"totalPv"`      // 累计浏览量
	UvGrowthRate *float64 `json:"uvGrowthRate"` // 访客增长率，昨日为 0 时为 null
	PvGrowthRate *float64 `json:"pvGrowthRate"` // 浏览量增长率，昨日为 0 时为 null
}

// UserSummaryVO 用户汇总VO
type UserSummaryVO struct {
	Total int64 `json:"total"` // 用户总数
}
