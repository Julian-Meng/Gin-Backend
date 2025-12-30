package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"backend/models"
)

// OpenAI / SiliconFlow / Ollama(OpenAI-Compatible) Chat Completions 响应结构
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`

	// 兼容一些实现可能返回 error 字段
	Error interface{} `json:"error,omitempty"`
}

// 默认 System Prompt
const defaultSystemPrompt = "你是企业人事管理系统的AI助手。请直接输出最终答案，不要输出任何推理过程、思考过程或类似“think”的内容。回答必须简洁、客观、面向业务。禁止输出<think>标签或任何推理文本。"

// GetAIResponse 调用 AI 接口（兼容本地 / SiliconFlow / OpenAI）
func GetAIResponse(req models.AIRequest) (models.AIResponse, error) {

	apiBase := strings.TrimSpace(os.Getenv("AI_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	modelID := strings.TrimSpace(os.Getenv("AI_MODEL_ID"))
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	systemPrompt := strings.TrimSpace(os.Getenv("AI_SYSTEM_PROMPT"))

	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}
	if provider == "" {
		// 默认为 ollama
		provider = "ollama"
	}

	// 基础校验
	if apiBase == "" || modelID == "" {
		return models.AIResponse{}, fmt.Errorf("AI environment variables not properly set: AI_BASE_URL / AI_MODEL_ID required")
	}

	// 统一：OpenAI Compatible 都是 /v1/chat/completions
	apiURL := strings.TrimRight(apiBase, "/") + "/chat/completions"

	// 构造请求体（OpenAI 标准格式）
	requestData := map[string]interface{}{
		"model":       modelID,
		"stream":      false,
		"temperature": 0.3,
		"top_p":       0.9,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": req.Message,
			},
		},
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to marshal request body: %v", err)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 只有非 ollama 才默认加鉴权头
	if provider != "ollama" && apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		// 对 ollama 本地给更友好的提示
		if provider == "ollama" {
			return models.AIResponse{}, fmt.Errorf("failed to call Ollama service: %v (请确认 Ollama 已启动，默认端口 11434)", err)
		}
		return models.AIResponse{}, fmt.Errorf("failed to call AI service: %v", err)
	}
	defer resp.Body.Close()

	// 非 200 时把 body 读出来，方便调试（比如 404 / 500）
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := strings.TrimSpace(string(bodyBytes))
		if bodyStr == "" {
			bodyStr = "(empty body)"
		}
		return models.AIResponse{}, fmt.Errorf("AI service returned status code: %d, body: %s", resp.StatusCode, bodyStr)
	}

	// 解析返回结果
	var rawResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to decode AI response: %v", err)
	}

	if len(rawResp.Choices) == 0 {
		return models.AIResponse{}, fmt.Errorf("AI response is empty")
	}

	content := rawResp.Choices[0].Message.Content
	content = sanitizeThink(content)

	// 统一封装为系统内部响应结构
	return models.AIResponse{
		Message: content,
	}, nil
}

// sanitizeThink：兜底清洗 <think>...</think> / 思考：... 之类的内容
func sanitizeThink(s string) string {
	if s == "" {
		return s
	}

	// 1) 清除 <think>...</think>
	reThinkTag := regexp.MustCompile(`(?is)<think>.*?</think>`)
	s = reThinkTag.ReplaceAllString(s, "")

	// 2) 清除类似 “思考：xxxx” 的段落（尽量保守，只删开头连续段）
	reThinkCn := regexp.MustCompile(`(?is)^\s*(思考|推理|分析)[:：].*?\n+`)
	s = reThinkCn.ReplaceAllString(s, "")

	// 3) 清理前后空白和多余空行
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")

	return s
}
