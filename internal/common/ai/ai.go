// Package ai 提供 OpenAI 兼容的 AI 对话调用
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"youlai-gin/internal/common/config"
	"youlai-gin/pkg/errs"
)

// Chat 调用 AI 对话接口，返回去掉代码围栏的文本内容
func Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	cfg := config.Cfg.AI
	if !cfg.Enabled || cfg.ApiKey == "" {
		return "", errs.BadRequest("AI 功能未开启，请配置 ai 段与 apiKey")
	}

	body, err := json.Marshal(map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return "", errs.SystemError("组装 AI 请求失败")
	}

	url := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", errs.SystemError("构建 AI 请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errs.ThirdPartyError("AI 服务调用失败")
	}
	defer resp.Body.Close()

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil || len(parsed.Choices) == 0 {
		return "", errs.ThirdPartyError("AI 返回内容解析失败")
	}

	return stripCodeFence(parsed.Choices[0].Message.Content), nil
}

// ChatJSON 调用 AI 对话接口并把返回内容解析到 v
func ChatJSON(ctx context.Context, systemPrompt, userPrompt string, v any) error {
	content, err := Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(content), v); err != nil {
		return errs.BadRequest("AI 返回内容非合法 JSON，请重试")
	}
	return nil
}

// stripCodeFence 去掉模型输出里可能包裹的 json 代码围栏
func stripCodeFence(content string) string {
	text := strings.TrimSpace(content)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	start := strings.Index(text, "\n")
	end := strings.LastIndex(text, "```")
	if start == -1 || end <= start {
		return text
	}
	return strings.TrimSpace(text[start+1 : end])
}
