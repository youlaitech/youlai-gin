package model

import "youlai-gin/pkg/types"

// AiAssistantRecordVO AI 助手行为记录 VO
// 对齐 youlai-boot AiAssistantRecordVO
// 仅用于记录分页列表返回
type AiAssistantRecordVO struct {
	ID                types.BigInt  `json:"id"`
	UserID            types.BigInt  `json:"userId"`
	Username          string        `json:"username"`
	OriginalCommand   string        `json:"originalCommand"`
	AiProvider        string        `json:"aiProvider"`
	AiModel           string        `json:"aiModel"`
	ParseStatus       int           `json:"parseStatus"`
	FunctionCalls     string        `json:"functionCalls"`
	Explanation       string        `json:"explanation"`
	Confidence        *float64      `json:"confidence"`
	ParseErrorMessage string        `json:"parseErrorMessage"`
	InputTokens       *int          `json:"inputTokens"`
	OutputTokens      *int          `json:"outputTokens"`
	ParseDurationMs   *int          `json:"parseDurationMs"`
	FunctionName      string        `json:"functionName"`
	FunctionArguments string        `json:"functionArguments"`
	ExecuteStatus     *int          `json:"executeStatus"`
	ExecuteErrorMsg   string        `gorm:"column:execute_error_message" json:"executeErrorMessage"`
	IPAddress         string        `json:"ipAddress"`
	CreateTime        types.LocalTime `json:"createTime"`
	UpdateTime        types.LocalTime `json:"updateTime"`
}
