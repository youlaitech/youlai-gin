package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"youlai-gin/internal/platform/ai/model"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/ai"
	appConfig "youlai-gin/pkg/config"
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

const systemPrompt = `你是一个智能的企业操作助手，需要将用户的自然语言命令解析成标准的函数调用。
请返回严格的 JSON 格式，包含字段：
- success: boolean
- explanation: string
- confidence: number (0-1)
- error: string
- provider: string
- model: string
- functionCalls: 数组，每个元素包含 name、description、arguments(对象)
当无法识别命令时，success=false，并给出 error。`

var (
	llmOnce sync.Once
	llmInst *openai.LLM
	llmErr  error
)

// ParseCommand 解析自然语言命令
func ParseCommand(ctx context.Context, req *model.AiParseRequest, userID int64, username string, ip string) (*model.AiParseResponse, error) {
	if req == nil || strings.TrimSpace(req.Command) == "" {
		return &model.AiParseResponse{Success: false, Error: "命令不能为空"}, nil
	}

	startTime := time.Now()
	record := &model.AiAssistantRecord{
		UserID:          types.BigInt(userID),
		Username:        username,
		OriginalCommand: strings.TrimSpace(req.Command),
		IPAddress:       ip,
		AiProvider:      getProvider(),
		AiModel:         getModel(),
	}

	payload := buildUserPrompt(req)
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, payload),
	}

	llmClient, err := getLLM()
	if err != nil {
		record.ParseStatus = 0
		record.ParseErrorMessage = err.Error()
		saveRecord(record, startTime)
		return nil, err
	}

	resp, err := llmClient.GenerateContent(ctx, messages, llms.WithJSONMode())
	if err != nil {
		record.ParseStatus = 0
		record.ParseErrorMessage = err.Error()
		saveRecord(record, startTime)
		return nil, err
	}
	if len(resp.Choices) == 0 {
		record.ParseStatus = 0
		record.ParseErrorMessage = "AI 返回内容为空"
		saveRecord(record, startTime)
		return nil, errs.SystemError("AI 返回内容为空")
	}

	raw := strings.TrimSpace(resp.Choices[0].Content)
	parsed, err := parseAIResponse(raw)
	if err != nil {
		record.ParseStatus = 0
		record.ParseErrorMessage = err.Error()
		record.FunctionCalls = "[]"
		record.Explanation = ""
		record.Confidence = nil
		saveRecord(record, startTime)
		return nil, err
	}

	record.AiProvider = defaultString(parsed.Provider, record.AiProvider)
	record.AiModel = defaultString(parsed.Model, record.AiModel)
	if parsed.Success {
		record.ParseStatus = 1
	} else {
		record.ParseStatus = 0
	}
	record.FunctionCalls = toJSONString(parsed.FunctionCalls)
	record.Explanation = parsed.Explanation
	record.Confidence = parsed.Confidence
	if !parsed.Success {
		record.ParseErrorMessage = defaultString(parsed.Error, "解析失败")
	}
	saveRecord(record, startTime)

	return &model.AiParseResponse{
		ParseLogID:    fmt.Sprintf("%d", record.ID),
		Success:       parsed.Success,
		FunctionCalls: parsed.FunctionCalls,
		Explanation:   parsed.Explanation,
		Confidence:    parsed.Confidence,
		Error:         parsed.Error,
		Provider:      record.AiProvider,
		Model:         record.AiModel,
		RawResponse:   raw,
	}, nil
}

// ExecuteCommand 执行解析后的命令
func ExecuteCommand(req *model.AiExecuteRequest, userID int64, username string, ip string) (*model.AiExecuteResponse, error) {
	if req == nil || req.FunctionCall.Name == "" {
		return nil, errs.BadRequest("函数调用不能为空")
	}

	record, err := loadOrCreateRecord(req, userID, username, ip)
	if err != nil {
		return nil, err
	}

	record.FunctionName = req.FunctionCall.Name
	record.FunctionArguments = toJSONString(req.FunctionCall.Arguments)
	execStatus := 0
	record.ExecuteStatus = &execStatus

	result, err := executeFunction(&req.FunctionCall)
	if err != nil {
		failedStatus := -1
		record.ExecuteStatus = &failedStatus
		record.ExecuteErrorMsg = err.Error()
		updateRecord(record)
		return &model.AiExecuteResponse{Success: false, Error: err.Error(), RecordID: fmt.Sprintf("%d", record.ID)}, nil
	}

	successStatus := 1
	record.ExecuteStatus = &successStatus
	record.ExecuteErrorMsg = ""
	updateRecord(record)

	return &model.AiExecuteResponse{
		Success:  true,
		Data:     result,
		Message:  "执行成功",
		RecordID: fmt.Sprintf("%d", record.ID),
	}, nil
}

// buildUserPrompt 拼接用户上下文
func buildUserPrompt(req *model.AiParseRequest) string {
	payload := map[string]interface{}{
		"command":          req.Command,
		"currentRoute":     req.CurrentRoute,
		"currentComponent": req.CurrentComponent,
		"context":          req.Context,
		"availableFunctions": []map[string]interface{}{
			{
				"name":               "updateUserNickname",
				"description":        "根据用户名更新用户昵称",
				"requiredParameters": []string{"username", "nickname"},
			},
		},
	}
	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return fmt.Sprintf("请根据以下上下文识别用户意图，并输出符合系统提示要求的 JSON：\n%s", string(bytes))
}

// parseAIResponse 解析 AI 返回的 JSON
func parseAIResponse(raw string) (*model.AiParseResponse, error) {
	if raw == "" {
		return nil, errs.SystemError("AI 返回内容为空")
	}

	var parsed model.AiParseResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, errs.SystemError("AI 响应解析失败")
	}
	return &parsed, nil
}

// executeFunction 执行函数调用
func executeFunction(call *model.AiFunctionCall) (interface{}, error) {
	switch call.Name {
	case "updateUserNickname":
		return executeUpdateUserNickname(call.Arguments)
	default:
		return nil, errs.BadRequest("不支持的函数: " + call.Name)
	}
}

// executeUpdateUserNickname 更新用户昵称
func executeUpdateUserNickname(args map[string]interface{}) (interface{}, error) {
	username := getStringArg(args, "username")
	nickname := getStringArg(args, "nickname")
	if username == "" || nickname == "" {
		return nil, errs.BadRequest("用户名或昵称不能为空")
	}

	user, err := userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		return nil, errs.NotFound("用户不存在")
	}

	user.Nickname = nickname
	if err := userRepo.UpdateUser(user); err != nil {
		return nil, errs.SystemError("更新用户昵称失败")
	}

	return map[string]interface{}{
		"username": username,
		"nickname": nickname,
		"message":  "用户昵称更新成功",
	}, nil
}

// loadOrCreateRecord 获取或创建执行记录
func loadOrCreateRecord(req *model.AiExecuteRequest, userID int64, username string, ip string) (*model.AiAssistantRecord, error) {
	if req.ParseLogID != "" {
		id, err := strconv.ParseInt(req.ParseLogID, 10, 64)
		if err != nil {
			return nil, errs.BadRequest("解析记录ID无效")
		}
		var record model.AiAssistantRecord
		if err := database.DB.First(&record, id).Error; err != nil {
			return nil, errs.NotFound("解析记录不存在")
		}
		return &record, nil
	}

	record := &model.AiAssistantRecord{
		UserID:          types.BigInt(userID),
		Username:        username,
		OriginalCommand: req.OriginalCommand,
		IPAddress:       ip,
		AiProvider:      getProvider(),
		AiModel:         getModel(),
	}
	if err := database.DB.Create(record).Error; err != nil {
		return nil, errs.SystemError("创建执行记录失败")
	}
	return record, nil
}

// saveRecord 保存解析记录
func saveRecord(record *model.AiAssistantRecord, start time.Time) {
	duration := int(time.Since(start).Milliseconds())
	record.ParseDurationMs = &duration
	_ = database.DB.Create(record).Error
}

// updateRecord 更新执行记录
func updateRecord(record *model.AiAssistantRecord) {
	_ = database.DB.Model(record).Updates(record).Error
}

// getLLM 获取 LLM 客户端
func getLLM() (*openai.LLM, error) {
	llmOnce.Do(func() {
		cfg := getAIConfig()
		baseURL := normalizeBaseURL(cfg.BaseURL)
		httpClient := &http.Client{Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond}
		llmInst, llmErr = openai.New(
			openai.WithBaseURL(baseURL),
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.Model),
			openai.WithHTTPClient(httpClient),
		)
	})
	return llmInst, llmErr
}

// normalizeBaseURL 处理兼容模式地址
func normalizeBaseURL(baseURL string) string {
	value := strings.TrimRight(baseURL, "/")
	if value == "" {
		return value
	}
	if strings.HasSuffix(value, "/v1") {
		return value
	}
	return value + "/v1"
}

// getAIConfig 读取 AI 配置
func getAIConfig() ai.Config {
	if appConfig.Cfg == nil {
		return ai.Config{}
	}
	return appConfig.Cfg.AI
}

// getProvider 获取配置的供应商
func getProvider() string {
	cfg := getAIConfig()
	if cfg.Provider == "" {
		return "qwen"
	}
	return cfg.Provider
}

// getModel 获取配置模型
func getModel() string {
	cfg := getAIConfig()
	if cfg.Model == "" {
		return "qwen-plus"
	}
	return cfg.Model
}

// toJSONString 转换 JSON 字符串
func toJSONString(data interface{}) string {
	if data == nil {
		return "[]"
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

// getStringArg 读取字符串参数
func getStringArg(args map[string]interface{}, key string) string {
	if args == nil {
		return ""
	}
	val, ok := args[key]
	if !ok || val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return strings.TrimSpace(str)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", val))
}

// defaultString 默认值处理
func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
