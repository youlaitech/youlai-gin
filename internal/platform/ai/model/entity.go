package model

import "youlai-gin/pkg/types"

// AiAssistantRecord AI 助手命令记录
// 对应表 ai_assistant_record
type AiAssistantRecord struct {
	ID                types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            types.BigInt `gorm:"column:user_id" json:"userId"`
	Username          string       `gorm:"column:username" json:"username"`
	OriginalCommand   string       `gorm:"column:original_command;type:text" json:"originalCommand"`
	AiProvider        string       `gorm:"column:ai_provider" json:"aiProvider"`
	AiModel           string       `gorm:"column:ai_model" json:"aiModel"`
	ParseStatus       int          `gorm:"column:parse_status" json:"parseStatus"`
	FunctionCalls     string       `gorm:"column:function_calls;type:text" json:"functionCalls"`
	Explanation       string       `gorm:"column:explanation" json:"explanation"`
	Confidence        *float64     `gorm:"column:confidence" json:"confidence"`
	ParseErrorMessage string       `gorm:"column:parse_error_message;type:text" json:"parseErrorMessage"`
	InputTokens       *int         `gorm:"column:input_tokens" json:"inputTokens"`
	OutputTokens      *int         `gorm:"column:output_tokens" json:"outputTokens"`
	ParseDurationMs   *int         `gorm:"column:parse_duration_ms" json:"parseDurationMs"`
	FunctionName      string       `gorm:"column:function_name" json:"functionName"`
	FunctionArguments string       `gorm:"column:function_arguments;type:text" json:"functionArguments"`
	ExecuteStatus     *int         `gorm:"column:execute_status" json:"executeStatus"`
	ExecuteErrorMsg   string       `gorm:"column:execute_error_message;type:text" json:"executeErrorMessage"`
	IPAddress         string       `gorm:"column:ip_address" json:"ipAddress"`
	CreateTime        string       `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime        string       `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (AiAssistantRecord) TableName() string {
	return "ai_assistant_record"
}
