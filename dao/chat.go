package dao

import (
	"backend/db"
	"backend/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OpenAI / SiliconFlow / Ollama(OpenAI-Compatible) Chat Completions 响应结构。
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`

	Error interface{} `json:"error,omitempty"`
}

const defaultSystemPrompt = "你是企业人事管理系统的AI助手。请直接输出最终答案，不要输出任何推理过程、思考过程或类似“think”的内容。回答必须简洁、客观、面向业务。禁止输出<think>标签或任何推理文本。"

// GetAIResponse 调用 AI 服务（OpenAI Compatible 协议）。
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
		provider = "ollama"
	}

	if apiBase == "" || modelID == "" {
		return models.AIResponse{}, fmt.Errorf("AI environment variables not properly set: AI_BASE_URL / AI_MODEL_ID required")
	}

	apiURL := strings.TrimRight(apiBase, "/") + "/chat/completions"

	requestData := map[string]interface{}{
		"model":       modelID,
		"stream":      false,
		"temperature": 0.3,
		"top_p":       0.9,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Message},
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

	if provider != "ollama" && apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		if provider == "ollama" {
			return models.AIResponse{}, fmt.Errorf("failed to call Ollama service: %v (请确认 Ollama 已启动，默认端口 11434)", err)
		}
		return models.AIResponse{}, fmt.Errorf("failed to call AI service: %v", err)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := strings.TrimSpace(string(bodyBytes))
		if bodyStr == "" {
			bodyStr = "(empty body)"
		}
		return models.AIResponse{}, fmt.Errorf("AI service returned status code: %d, body: %s", resp.StatusCode, bodyStr)
	}

	var rawResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return models.AIResponse{}, fmt.Errorf("failed to decode AI response: %v", err)
	}
	if len(rawResp.Choices) == 0 {
		return models.AIResponse{}, fmt.Errorf("AI response is empty")
	}

	content := sanitizeThink(rawResp.Choices[0].Message.Content)
	return models.AIResponse{Message: content}, nil
}

// sanitizeThink 清洗模型可能返回的思考内容。
func sanitizeThink(s string) string {
	if s == "" {
		return s
	}

	reThinkTag := regexp.MustCompile(`(?is)<think>.*?</think>`)
	s = reThinkTag.ReplaceAllString(s, "")

	reThinkCn := regexp.MustCompile(`(?is)^\s*(思考|推理|分析)[:：].*?\n+`)
	s = reThinkCn.ReplaceAllString(s, "")

	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	return s
}

// CreateUserChatMessage 创建用户消息，并在管理员在线时进行会话粘性分配。
func CreateUserChatMessage(userEmpID, content string, onlineAdminEmpIDs []string) (models.ChatSession, models.ChatMessage, bool, error) {
	dbConn := db.GetDB()
	now := time.Now()
	hasOnlineAdmin := len(onlineAdminEmpIDs) > 0

	var outSession models.ChatSession
	var outMessage models.ChatMessage

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		session, err := getOrCreateOpenSessionForUpdate(tx, userEmpID, now)
		if err != nil {
			return err
		}

		if session.AssignedAdminEmpID == nil && hasOnlineAdmin {
			picked, pickErr := pickRandomLeastLoadedAdmin(tx, onlineAdminEmpIDs)
			if pickErr != nil {
				return pickErr
			}
			if picked != "" {
				session.AssignedAdminEmpID = &picked
				session.Status = models.ChatSessionStatusOpen
			}
		}

		if session.AssignedAdminEmpID == nil {
			session.Status = models.ChatSessionStatusWaiting
		}

		session.LastMessageAt = now
		if err := tx.Save(&session).Error; err != nil {
			return err
		}

		msg := models.ChatMessage{
			SessionID:     session.ID,
			SenderType:    models.ChatMessageSenderUser,
			SenderEmpID:   userEmpID,
			Content:       content,
			MsgType:       models.ChatMessageTypeText,
			DeliveryState: models.ChatDeliveryStateQueued,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		outSession = session
		outMessage = msg
		return nil
	})

	return outSession, outMessage, hasOnlineAdmin, err
}

// CreateAdminChatMessage 管理员向会话发送消息。
func CreateAdminChatMessage(sessionID uint, adminEmpID, content string) (models.ChatSession, models.ChatMessage, error) {
	dbConn := db.GetDB()
	now := time.Now()

	var outSession models.ChatSession
	var outMessage models.ChatMessage

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var session models.ChatSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&session, sessionID).Error; err != nil {
			return err
		}

		// 会话若未绑定，管理员首条消息可直接接管。
		if session.AssignedAdminEmpID == nil {
			session.AssignedAdminEmpID = &adminEmpID
		}
		session.Status = models.ChatSessionStatusOpen
		session.LastMessageAt = now
		session.Source = normalizeSessionSource(session.Source, models.ChatSessionSourceHuman)

		if err := tx.Save(&session).Error; err != nil {
			return err
		}

		msg := models.ChatMessage{
			SessionID:     session.ID,
			SenderType:    models.ChatMessageSenderAdmin,
			SenderEmpID:   adminEmpID,
			Content:       content,
			MsgType:       models.ChatMessageTypeText,
			DeliveryState: models.ChatDeliveryStateQueued,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		outSession = session
		outMessage = msg
		return nil
	})

	return outSession, outMessage, err
}

// CreateAIFallbackMessage 写入 AI 兜底回复，供管理员后续追溯。
func CreateAIFallbackMessage(sessionID uint, content, notice string) (models.ChatSession, models.ChatMessage, error) {
	dbConn := db.GetDB()
	now := time.Now()

	var outSession models.ChatSession
	var outMessage models.ChatMessage

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var session models.ChatSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, sessionID).Error; err != nil {
			return err
		}

		session.LastMessageAt = now
		session.Source = normalizeSessionSource(session.Source, models.ChatSessionSourceAIFallback)
		if err := tx.Save(&session).Error; err != nil {
			return err
		}

		msg := models.ChatMessage{
			SessionID:     session.ID,
			SenderType:    models.ChatMessageSenderAI,
			Content:       content,
			MsgType:       models.ChatMessageTypeText,
			DeliveryState: models.ChatDeliveryStateQueued,
			AIFlag:        true,
			AINotice:      notice,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		outSession = session
		outMessage = msg
		return nil
	})

	return outSession, outMessage, err
}

// ClaimWaitingSessionsForAdmin 管理员上线后认领待处理会话（并发安全，幂等）。
func ClaimWaitingSessionsForAdmin(adminEmpID string, limit int) ([]models.ChatSession, error) {
	if limit <= 0 {
		limit = 20
	}

	dbConn := db.GetDB()
	claimed := make([]models.ChatSession, 0, limit)

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var candidates []models.ChatSession
		if err := tx.Where("status = ? AND assigned_admin_emp_id IS NULL", models.ChatSessionStatusWaiting).
			Order("last_message_at ASC").
			Limit(limit).
			Find(&candidates).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, c := range candidates {
			res := tx.Model(&models.ChatSession{}).
				Where("id = ? AND status = ? AND assigned_admin_emp_id IS NULL", c.ID, models.ChatSessionStatusWaiting).
				Updates(map[string]interface{}{
					"assigned_admin_emp_id": adminEmpID,
					"status":                models.ChatSessionStatusOpen,
					"updated_at":            now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				continue
			}

			var claimedSession models.ChatSession
			if err := tx.First(&claimedSession, c.ID).Error; err != nil {
				return err
			}
			claimed = append(claimed, claimedSession)
		}

		return nil
	})

	return claimed, err
}

// ListSessionsForAdmin 查询管理员可处理的会话。
func ListSessionsForAdmin(adminEmpID string, limit int) ([]models.ChatSession, error) {
	if limit <= 0 {
		limit = 100
	}

	dbConn := db.GetDB()
	var sessions []models.ChatSession
	err := dbConn.Where("assigned_admin_emp_id = ?", adminEmpID).
		Order("last_message_at DESC").
		Limit(limit).
		Find(&sessions).Error
	return sessions, err
}

// ListSessionsForUser 查询用户自己的会话（MVP 默认一人一个活跃会话）。
func ListSessionsForUser(userEmpID string, limit int) ([]models.ChatSession, error) {
	if limit <= 0 {
		limit = 20
	}
	dbConn := db.GetDB()
	var sessions []models.ChatSession
	err := dbConn.Where("user_emp_id = ?", userEmpID).
		Order("last_message_at DESC").
		Limit(limit).
		Find(&sessions).Error
	return sessions, err
}

// ListMessagesBySession 查询会话消息。
func ListMessagesBySession(sessionID uint, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 200
	}
	dbConn := db.GetDB()
	var messages []models.ChatMessage
	err := dbConn.Where("session_id = ?", sessionID).
		Order("id ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetSessionByID 查询会话。
func GetSessionByID(sessionID uint) (models.ChatSession, error) {
	dbConn := db.GetDB()
	var session models.ChatSession
	err := dbConn.First(&session, sessionID).Error
	return session, err
}

// ClaimSessionByID 管理员手动接管会话（未分配或已归属于本人时成功）。
func ClaimSessionByID(sessionID uint, adminEmpID string) (models.ChatSession, bool, error) {
	dbConn := db.GetDB()
	var out models.ChatSession

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var session models.ChatSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, sessionID).Error; err != nil {
			return err
		}

		if session.AssignedAdminEmpID != nil && *session.AssignedAdminEmpID != adminEmpID {
			out = session
			return nil
		}

		now := time.Now()
		res := tx.Model(&models.ChatSession{}).
			Where("id = ? AND (assigned_admin_emp_id IS NULL OR assigned_admin_emp_id = ?)", sessionID, adminEmpID).
			Updates(map[string]interface{}{
				"assigned_admin_emp_id": adminEmpID,
				"status":                models.ChatSessionStatusOpen,
				"updated_at":            now,
			})
		if res.Error != nil {
			return res.Error
		}

		if err := tx.First(&out, sessionID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return models.ChatSession{}, false, err
	}

	ok := out.AssignedAdminEmpID != nil && *out.AssignedAdminEmpID == adminEmpID
	return out, ok, nil
}

// CanUserAccessSession 验证用户是否可访问会话。
func CanUserAccessSession(sessionID uint, userEmpID string) (bool, error) {
	session, err := GetSessionByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return session.UserEmpID == userEmpID, nil
}

// CanAdminAccessSession 验证管理员是否可访问会话。
func CanAdminAccessSession(sessionID uint, adminEmpID string) (bool, error) {
	session, err := GetSessionByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if session.AssignedAdminEmpID == nil {
		return true, nil
	}
	return *session.AssignedAdminEmpID == adminEmpID, nil
}

// TouchMessagesAsDispatched 将 queued 状态标记为 dispatched（用于已推送到连接层）。
func TouchMessagesAsDispatched(messageIDs []uint) error {
	if len(messageIDs) == 0 {
		return nil
	}
	dbConn := db.GetDB()
	now := time.Now()
	return dbConn.Model(&models.ChatMessage{}).
		Where("id IN ? AND delivery_state = ?", messageIDs, models.ChatDeliveryStateQueued).
		Updates(map[string]interface{}{
			"delivery_state": models.ChatDeliveryStateDispatched,
			"delivered_at":   &now,
		}).Error
}

// GetQueuedMessagesBySession 查询会话中待派发消息。
func GetQueuedMessagesBySession(sessionID uint) ([]models.ChatMessage, error) {
	dbConn := db.GetDB()
	var messages []models.ChatMessage
	err := dbConn.Where("session_id = ? AND delivery_state = ?", sessionID, models.ChatDeliveryStateQueued).
		Order("id ASC").
		Find(&messages).Error
	return messages, err
}

// GetLatestAIFallbackMessage 查询会话最近一条 AI 兜底消息。
func GetLatestAIFallbackMessage(sessionID uint) (models.ChatMessage, error) {
	dbConn := db.GetDB()
	var msg models.ChatMessage
	err := dbConn.Where("session_id = ? AND sender_type = ? AND ai_flag = ?", sessionID, models.ChatMessageSenderAI, true).
		Order("id DESC").
		First(&msg).Error
	return msg, err
}

func getOrCreateOpenSessionForUpdate(tx *gorm.DB, userEmpID string, now time.Time) (models.ChatSession, error) {
	var session models.ChatSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_emp_id = ? AND status <> ?", userEmpID, models.ChatSessionStatusClosed).
		Order("updated_at DESC").
		First(&session).Error

	if err == nil {
		return session, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ChatSession{}, err
	}

	session = models.ChatSession{
		UserEmpID:     userEmpID,
		Status:        models.ChatSessionStatusWaiting,
		Source:        models.ChatSessionSourceHuman,
		LastMessageAt: now,
	}
	if err := tx.Create(&session).Error; err != nil {
		return models.ChatSession{}, err
	}
	return session, nil
}

func pickRandomLeastLoadedAdmin(tx *gorm.DB, onlineAdminEmpIDs []string) (string, error) {
	if len(onlineAdminEmpIDs) == 0 {
		return "", nil
	}

	type row struct {
		AdminEmpID string
		Cnt        int64
	}

	rows := make([]row, 0)
	if err := tx.Model(&models.ChatSession{}).
		Select("assigned_admin_emp_id as admin_emp_id, COUNT(*) as cnt").
		Where("assigned_admin_emp_id IN ? AND status <> ?", onlineAdminEmpIDs, models.ChatSessionStatusClosed).
		Group("assigned_admin_emp_id").
		Scan(&rows).Error; err != nil {
		return "", err
	}

	countMap := make(map[string]int64, len(onlineAdminEmpIDs))
	for _, id := range onlineAdminEmpIDs {
		countMap[id] = 0
	}
	for _, r := range rows {
		countMap[r.AdminEmpID] = r.Cnt
	}

	minCnt := int64(1<<62 - 1)
	for _, c := range countMap {
		if c < minCnt {
			minCnt = c
		}
	}

	candidates := make([]string, 0)
	for adminID, c := range countMap {
		if c == minCnt {
			candidates = append(candidates, adminID)
		}
	}

	if len(candidates) == 0 {
		return "", nil
	}
	return candidates[rand.Intn(len(candidates))], nil
}

func normalizeSessionSource(oldSource, incoming string) string {
	if oldSource == "" || oldSource == incoming {
		return incoming
	}
	if oldSource == models.ChatSessionSourceMixed {
		return oldSource
	}
	return models.ChatSessionSourceMixed
}
