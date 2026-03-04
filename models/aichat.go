package models

// AIRequest AI请求结构体
type AIRequest struct {
	Message   string `json:"message"`    // 用户输入的消息
	SessionID string `json:"session_id"` // 用于维持会话状态
}

// AIResponse AI响应结构体
type AIResponse struct {
	Message string `json:"message"` // AI返回的消息
}
