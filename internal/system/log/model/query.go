package model

import "youlai-gin/pkg/common"

// LogPageQuery 日志分页查询
type LogQuery struct {
	common.BaseQuery
	Module    string `form:"module"`    // 操作模块
	Username  string `form:"username"`  // 操作用户
	Status    *int   `form:"status"`    // 状态码
	StartTime string `form:"startTime"` // 开始时间
	EndTime   string `form:"endTime"`   // 结束时间
}
