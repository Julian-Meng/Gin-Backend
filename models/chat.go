package models

import "time"

const (
	ChatSessionStatusOpen       = "open"
	ChatSessionStatusWaiting    = "waiting_admin"
	ChatSessionStatusClosed     = "closed"
	ChatSessionSourceHuman      = "human"
	ChatSessionSourceAIFallback = "ai_fallback"
	ChatSessionSourceMixed      = "mixed"
	ChatMessageTypeText         = "text"
	ChatMessageTypeEvent        = "event"
	ChatMessageSenderUser       = "user"
	ChatMessageSenderAdmin      = "admin"
	ChatMessageSenderAI         = "ai"
	ChatMessageSenderSystem     = "system"
	ChatDeliveryStateQueued     = "queued"
	ChatDeliveryStateDispatched = "dispatched"
	ChatDeliveryStateDelivered  = "delivered"
	ChatDeliveryStateRead       = "read"
	ChatDeliveryStateFailed     = "failed"
)

// ChatSession 用户与管理员会话。
type ChatSession struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserEmpID          string    `gorm:"size:32;index;not null" json:"user_emp_id"`
	AssignedAdminEmpID *string   `gorm:"size:32;index" json:"assigned_admin_emp_id,omitempty"`
	Status             string    `gorm:"size:32;index;not null;default:'waiting_admin'" json:"status"`
	Source             string    `gorm:"size:32;not null;default:'human'" json:"source"`
	LastMessageAt      time.Time `gorm:"index;not null" json:"last_message_at"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

func (ChatSession) TableName() string {
	return "CHAT_SESSION"
}

// ChatMessage 会话消息及投递状态。
type ChatMessage struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	SessionID     uint       `gorm:"index;not null" json:"session_id"`
	SenderType    string     `gorm:"size:16;index;not null" json:"sender_type"`
	SenderEmpID   string     `gorm:"size:32;index" json:"sender_emp_id,omitempty"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	MsgType       string     `gorm:"size:16;not null;default:'text'" json:"msg_type"`
	DeliveryState string     `gorm:"size:20;index;not null;default:'queued'" json:"delivery_state"`
	AIFlag        bool       `gorm:"index;not null;default:false" json:"ai_flag"`
	AINotice      string     `gorm:"size:255" json:"ai_notice,omitempty"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"update_at"`
}

func (ChatMessage) TableName() string {
	return "CHAT_MESSAGE"
}

// AIRequest AI 请求结构体。
type AIRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// AIResponse AI 响应结构体。
type AIResponse struct {
	Message string `json:"message"`
}

// UserSendChatRequest 用户发送消息请求。
type UserSendChatRequest struct {
	Content string `json:"content"`
}

// AdminSendChatRequest 管理员发送消息请求。
type AdminSendChatRequest struct {
	SessionID uint   `json:"session_id"`
	Content   string `json:"content"`
}

// AdminClaimRequest 管理员认领等待会话请求。
type AdminClaimRequest struct {
	Limit int `json:"limit"`
}
