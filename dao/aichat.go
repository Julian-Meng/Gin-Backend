package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"backend/models"
)

// SiliconFlow / OpenAI Chat Completions 响应结构）

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 调用 AI 接口

func GetAIResponse(req models.AIRequest) (models.AIResponse, error) {

	apiBase := os.Getenv("AI_BASE_URL")
	apiKey := os.Getenv("AI_API_KEY")
	modelID := os.Getenv("AI_MODEL_ID")

	if apiBase == "" || apiKey == "" || modelID == "" {
		return models.AIResponse{}, fmt.Errorf("AI environment variables not properly set")
	}

	apiURL := apiBase + "/chat/completions"

	// 构造请求体（OpenAI 标准格式）
	requestData := map[string]interface{}{
		"model": modelID,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是企业人事管理系统中的AI助手，负责对统计数据进行简要、客观的分析。",
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

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("X-API-Key", apiKey)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to call AI service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.AIResponse{}, fmt.Errorf("AI service returned status code: %d", resp.StatusCode)
	}

	// 解析返回结果
	var rawResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to decode AI response: %v", err)
	}

	if len(rawResp.Choices) == 0 {
		return models.AIResponse{}, fmt.Errorf("AI response is empty")
	}

	// 统一封装为系统内部响应结构
	return models.AIResponse{
		Message: rawResp.Choices[0].Message.Content,
	}, nil
}
