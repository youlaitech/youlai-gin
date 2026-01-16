package model

// AiParseRequest AI 命令解析请求
// 包含用户输入和上下文信息
type AiParseRequest struct {
	Command          string                 `json:"command"`
	CurrentRoute     string                 `json:"currentRoute"`
	CurrentComponent string                 `json:"currentComponent"`
	Context          map[string]interface{} `json:"context"`
}

// AiFunctionCall AI 函数调用信息
type AiFunctionCall struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   map[string]interface{} `json:"arguments"`
}

// AiParseResponse AI 命令解析响应
type AiParseResponse struct {
	ParseLogID    string           `json:"parseLogId"`
	Success       bool             `json:"success"`
	FunctionCalls []AiFunctionCall `json:"functionCalls"`
	Explanation   string           `json:"explanation"`
	Confidence    *float64         `json:"confidence"`
	Error         string           `json:"error"`
	Provider      string           `json:"provider"`
	Model         string           `json:"model"`
	RawResponse   string           `json:"rawResponse"`
}

// AiExecuteRequest AI 命令执行请求
type AiExecuteRequest struct {
	ParseLogID      string         `json:"parseLogId"`
	OriginalCommand string         `json:"originalCommand"`
	FunctionCall    AiFunctionCall `json:"functionCall"`
	ConfirmMode     string         `json:"confirmMode"`
	UserConfirmed   *bool          `json:"userConfirmed"`
	IdempotencyKey  string         `json:"idempotencyKey"`
	CurrentRoute    string         `json:"currentRoute"`
}

// AiExecuteResponse AI 命令执行响应
type AiExecuteResponse struct {
	Success              bool        `json:"success"`
	Data                 interface{} `json:"data"`
	Message              string      `json:"message"`
	AffectedRows         *int        `json:"affectedRows"`
	Error                string      `json:"error"`
	RecordID             string      `json:"recordId"`
	RequiresConfirmation *bool       `json:"requiresConfirmation"`
	ConfirmationPrompt   string      `json:"confirmationPrompt"`
}
