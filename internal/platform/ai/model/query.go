package model

import "youlai-gin/pkg/common"

// AiAssistantQuery AI 助手行为记录查询对象
// 对齐 youlai-boot AiAssistantQuery
// form 参数以查询字符串传递
// keywords: 原始命令/函数名称/用户名 关键字
// createTime: 创建时间范围
// executeStatus: 执行状态(0-待执行, 1-成功, -1-失败)
// parseStatus: 解析状态(0-失败, 1-成功)
type AiAssistantQuery struct {
	common.BaseQuery
	Keywords      string   `form:"keywords"`
	ExecuteStatus *int     `form:"executeStatus"`
	UserID        *int64   `form:"userId"`
	ParseStatus   *int     `form:"parseStatus"`
	CreateTime    []string `form:"createTime"`
	FunctionName  string   `form:"functionName"`
	AiProvider    string   `form:"aiProvider"`
	AiModel       string   `form:"aiModel"`
}
