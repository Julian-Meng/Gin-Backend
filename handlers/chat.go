package handlers

import (
	"backend/dao"
	"backend/middlewares"
	"backend/models"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const aiFallbackNotice = "当前为自动回复，管理员上线后会继续跟进。"

type wsClient struct {
	empID string
	role  string
	conn  *websocket.Conn
	mu    sync.Mutex
}

type chatHub struct {
	mu     sync.RWMutex
	admins map[string]map[*wsClient]struct{}
	users  map[string]map[*wsClient]struct{}
}

type chatPushPayload struct {
	Type      string              `json:"type"`
	SessionID uint                `json:"session_id,omitempty"`
	Session   *models.ChatSession `json:"session,omitempty"`
	Message   *models.ChatMessage `json:"message,omitempty"`
	Notice    string              `json:"notice,omitempty"`
	SentAt    time.Time           `json:"sent_at"`
}

var runtimeChatHub = newChatHub()

func newChatHub() *chatHub {
	return &chatHub{
		admins: make(map[string]map[*wsClient]struct{}),
		users:  make(map[string]map[*wsClient]struct{}),
	}
}

func (h *chatHub) addConnection(role, empID string, conn *websocket.Conn) *wsClient {
	client := &wsClient{empID: empID, role: role, conn: conn}

	h.mu.Lock()
	defer h.mu.Unlock()

	if role == "admin" {
		if h.admins[empID] == nil {
			h.admins[empID] = make(map[*wsClient]struct{})
		}
		h.admins[empID][client] = struct{}{}
		return client
	}

	if h.users[empID] == nil {
		h.users[empID] = make(map[*wsClient]struct{})
	}
	h.users[empID][client] = struct{}{}
	return client
}

func (h *chatHub) removeConnection(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.role == "admin" {
		if set, ok := h.admins[client.empID]; ok {
			delete(set, client)
			if len(set) == 0 {
				delete(h.admins, client.empID)
			}
		}
		return
	}

	if set, ok := h.users[client.empID]; ok {
		delete(set, client)
		if len(set) == 0 {
			delete(h.users, client.empID)
		}
	}
}

func (h *chatHub) getOnlineAdminEmpIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.admins))
	for adminID := range h.admins {
		ids = append(ids, adminID)
	}
	return ids
}

func (h *chatHub) isAdminOnline(adminEmpID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	set := h.admins[adminEmpID]
	return len(set) > 0
}

func (h *chatHub) pushToAdmin(adminEmpID string, payload chatPushPayload) bool {
	clients := h.copyClients("admin", adminEmpID)
	if len(clients) == 0 {
		return false
	}

	success := false
	for _, c := range clients {
		if c.writeJSON(payload) == nil {
			success = true
			continue
		}
		h.removeConnection(c)
		_ = c.conn.Close()
	}
	return success
}

func (h *chatHub) pushToUser(userEmpID string, payload chatPushPayload) bool {
	clients := h.copyClients("user", userEmpID)
	if len(clients) == 0 {
		return false
	}

	success := false
	for _, c := range clients {
		if c.writeJSON(payload) == nil {
			success = true
			continue
		}
		h.removeConnection(c)
		_ = c.conn.Close()
	}
	return success
}

func (h *chatHub) copyClients(role, empID string) []*wsClient {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var src map[*wsClient]struct{}
	if role == "admin" {
		src = h.admins[empID]
	} else {
		src = h.users[empID]
	}

	clients := make([]*wsClient, 0, len(src))
	for c := range src {
		clients = append(clients, c)
	}
	return clients
}

func (c *wsClient) writeJSON(payload chatPushPayload) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(payload)
}

func (c *wsClient) writePing() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second))
}

// ChatWithAI 兼容旧版 AI 聊天接口。
func ChatWithAI(c *gin.Context) {
	var req models.AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "default-session-" + time.Now().Format("20060102150405")
	}

	aiResponse, err := dao.GetAIResponse(models.AIRequest{
		Message:   req.Message,
		SessionID: sessionID,
	})
	if err != nil {
		log.Printf("AI ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aiResponse)
}

// ChatWS 建立聊天 WebSocket 连接（用户和管理员均可）。
func ChatWS(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipalForWS(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}

	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := runtimeChatHub.addConnection(principalRole, empID, conn)

	if principalRole == "admin" {
		claimed, _ := dao.ClaimWaitingSessionsForAdmin(empID, 20)
		for _, s := range claimed {
			runtimeChatHub.pushToAdmin(empID, chatPushPayload{
				Type:      "session_claimed",
				SessionID: s.ID,
				Session:   &s,
				SentAt:    time.Now(),
			})
			dispatchQueuedMessagesForSession(s.ID)
		}
	}

	defer func() {
		runtimeChatHub.removeConnection(client)
		_ = conn.Close()
	}()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	})

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	stopPing := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stopPing:
				return
			case <-ticker.C:
				if err := client.writePing(); err != nil {
					return
				}
			}
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			close(stopPing)
			<-done
			return
		}
	}
}

// UserSendChatMessage 用户发送消息。
func UserSendChatMessage(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "user" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅普通用户可调用该接口"})
		return
	}

	var req models.UserSendChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请求格式错误", "err": err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "消息内容不能为空"})
		return
	}

	onlineAdmins := runtimeChatHub.getOnlineAdminEmpIDs()
	session, userMsg, _, err := dao.CreateUserChatMessage(empID, req.Content, onlineAdmins)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "发送失败", "err": err.Error()})
		return
	}

	_ = dispatchQueuedMessagesForSession(session.ID)

	var aiMsg *models.ChatMessage
	if shouldUseAIFallback(session) {
		generated, aiErr := tryGenerateAIFallback(session, req.Content)
		if aiErr == nil && generated != nil {
			aiMsg = generated
			_ = dispatchQueuedMessagesForSession(session.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": gin.H{
			"session":      session,
			"user_message": userMsg,
			"ai_message":   aiMsg,
		},
	})
}

// AdminSendChatMessage 管理员发送消息。
func AdminSendChatMessage(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅管理员可调用该接口"})
		return
	}

	var req models.AdminSendChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请求格式错误", "err": err.Error()})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.SessionID == 0 || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "session_id 和 content 不能为空"})
		return
	}

	allowed, err := dao.CanAdminAccessSession(req.SessionID, empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "鉴权失败", "err": err.Error()})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "该会话已分配给其他管理员"})
		return
	}

	session, msg, err := dao.CreateAdminChatMessage(req.SessionID, empID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "发送失败", "err": err.Error()})
		return
	}

	_ = dispatchQueuedMessagesForSession(session.ID)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": gin.H{
			"session": session,
			"message": msg,
		},
	})
}

// AdminClaimWaitingSessions 管理员主动认领等待会话。
func AdminClaimWaitingSessions(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅管理员可调用该接口"})
		return
	}

	var req models.AdminClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body，使用默认值
		req.Limit = 20
	}

	claimed, err := dao.ClaimWaitingSessionsForAdmin(empID, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "认领失败", "err": err.Error()})
		return
	}

	for _, s := range claimed {
		_ = dispatchQueuedMessagesForSession(s.ID)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": claimed})
}

// AdminClaimSessionByID 管理员手动接管指定会话。
func AdminClaimSessionByID(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅管理员可调用该接口"})
		return
	}

	sessionID, err := parseSessionIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效会话ID"})
		return
	}

	session, okClaimed, err := dao.ClaimSessionByID(sessionID, empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "接管失败", "err": err.Error()})
		return
	}
	if !okClaimed {
		c.JSON(http.StatusConflict, gin.H{"code": 1, "msg": "会话已被其他管理员接管"})
		return
	}

	_ = dispatchQueuedMessagesForSession(session.ID)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": session})
}

// UserListChatSessions 用户查询自己的会话。
func UserListChatSessions(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "user" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅普通用户可调用该接口"})
		return
	}

	sessions, err := dao.ListSessionsForUser(empID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": sessions})
}

// AdminListChatSessions 管理员查询已分配会话。
func AdminListChatSessions(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}
	if principalRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅管理员可调用该接口"})
		return
	}

	sessions, err := dao.ListSessionsForAdmin(empID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": sessions})
}

// GetChatMessages 查询会话消息历史。
func GetChatMessages(c *gin.Context) {
	empID, principalRole, ok := currentChatPrincipal(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录或身份无效"})
		return
	}

	sessionID, err := parseSessionIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效会话ID"})
		return
	}

	if principalRole == "admin" {
		allowed, aErr := dao.CanAdminAccessSession(sessionID, empID)
		if aErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "鉴权失败", "err": aErr.Error()})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "无权访问该会话"})
			return
		}
	} else {
		allowed, aErr := dao.CanUserAccessSession(sessionID, empID)
		if aErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "鉴权失败", "err": aErr.Error()})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "无权访问该会话"})
			return
		}
	}

	messages, err := dao.ListMessagesBySession(sessionID, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": messages})
}

func dispatchQueuedMessagesForSession(sessionID uint) error {
	session, err := dao.GetSessionByID(sessionID)
	if err != nil {
		return err
	}

	queued, err := dao.GetQueuedMessagesBySession(sessionID)
	if err != nil {
		return err
	}

	toTouch := make([]uint, 0, len(queued))
	for _, msg := range queued {
		payload := chatPushPayload{
			Type:      "message",
			SessionID: session.ID,
			Session:   &session,
			Message:   &msg,
			SentAt:    time.Now(),
		}

		sent := false
		switch msg.SenderType {
		case models.ChatMessageSenderUser:
			if session.AssignedAdminEmpID != nil {
				sent = runtimeChatHub.pushToAdmin(*session.AssignedAdminEmpID, payload)
			}
		case models.ChatMessageSenderAdmin, models.ChatMessageSenderAI, models.ChatMessageSenderSystem:
			sent = runtimeChatHub.pushToUser(session.UserEmpID, payload)
		}

		if sent {
			toTouch = append(toTouch, msg.ID)
		}
	}

	return dao.TouchMessagesAsDispatched(toTouch)
}

func shouldUseAIFallback(session models.ChatSession) bool {
	if !getEnvBoolWithDefault("CHAT_AI_FALLBACK_ENABLED", true) {
		return false
	}

	if session.AssignedAdminEmpID == nil {
		return true
	}

	return !runtimeChatHub.isAdminOnline(*session.AssignedAdminEmpID)
}

func tryGenerateAIFallback(session models.ChatSession, userContent string) (*models.ChatMessage, error) {
	minGapSeconds := getEnvIntWithDefault("CHAT_AI_FALLBACK_MIN_GAP_SECONDS", 30)
	latestAI, err := dao.GetLatestAIFallbackMessage(session.ID)
	if err == nil {
		if time.Since(latestAI.CreatedAt) < time.Duration(minGapSeconds)*time.Second {
			return nil, nil
		}
	}
	if err != nil && !errorsIsRecordNotFound(err) {
		return nil, err
	}

	aiResp, err := dao.GetAIResponse(models.AIRequest{
		SessionID: strconv.FormatUint(uint64(session.ID), 10),
		Message:   buildOfflineAIPrompt(userContent),
	})
	if err != nil {
		return nil, err
	}

	_, aiMsg, err := dao.CreateAIFallbackMessage(session.ID, aiResp.Message, aiFallbackNotice)
	if err != nil {
		return nil, err
	}

	return &aiMsg, nil
}

func buildOfflineAIPrompt(userContent string) string {
	return "你是企业人事系统的离线客服助手。你只能回答：制度说明、流程指引、系统操作说明、公告/考勤/请假等通用解释。" +
		"禁止承诺或执行任何写操作（审批、改考勤、改权限、改账户、导出隐私）。如果涉及敏感操作或信息不足，请明确回复需要管理员处理。" +
		"最后必须追加一句：当前为自动回复，管理员上线后会继续跟进。用户问题：" + userContent
}

func currentChatPrincipal(c *gin.Context) (string, string, bool) {
	roleAny, roleExists := c.Get("role")
	if !roleExists {
		return "", "", false
	}
	role, _ := roleAny.(string)

	empIDAny, empExists := c.Get("emp_id")
	empID, _ := empIDAny.(string)

	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)

	return principalFromClaims(strings.TrimSpace(username), strings.TrimSpace(empID), role, empExists)
}

func currentChatPrincipalForWS(c *gin.Context) (string, string, bool) {
	if empID, principalRole, ok := currentChatPrincipal(c); ok {
		return empID, principalRole, true
	}

	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
	}
	if token == "" {
		return "", "", false
	}

	claims, err := middlewares.ParseToken(token)
	if err != nil {
		return "", "", false
	}

	hasEmpID := strings.TrimSpace(claims.EmpID) != ""
	return principalFromClaims(strings.TrimSpace(claims.Username), strings.TrimSpace(claims.EmpID), strings.TrimSpace(claims.Role), hasEmpID)
}

func principalFromClaims(username, empID, role string, hasEmpID bool) (string, string, bool) {
	role = strings.TrimSpace(role)

	if role == "admin" {
		if !hasEmpID || empID == "" {
			return "", "", false
		}
		return empID, "admin", true
	}

	if role == "superadmin" {
		if hasEmpID && empID != "" {
			return empID, "admin", true
		}
		if username == "" {
			return "", "", false
		}
		return "superadmin:" + username, "admin", true
	}

	if role == "staff" {
		if !hasEmpID || empID == "" {
			return "", "", false
		}
		return empID, "user", true
	}

	return "", "", false
}

func parseSessionIDParam(raw string) (uint, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || v == 0 {
		return 0, err
	}
	return uint(v), nil
}

func getEnvBoolWithDefault(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getEnvIntWithDefault(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func errorsIsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
