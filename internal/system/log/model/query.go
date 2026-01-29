package model

import "youlai-gin/pkg/common"

// LogPageQuery 日志分页查询
type LogQuery struct {
	common.BaseQuery
	Keywords   string   `form:"keywords"`   // 关键字(日志内容/请求路径/请求方法/地区/浏览器/终端系统)
	CreateTime []string `form:"createTime"` // 操作时间范围
}
