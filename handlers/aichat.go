package handlers

import (
	"backend/dao"
	"backend/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ChatWithAI 处理AI聊天请求
func ChatWithAI(c *gin.Context) {
	var req models.AIRequest

	// 解析请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 为每个会话生成一个临时的session_id (可以用JWT中的emp_id)
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "default-session-" + time.Now().Format("20060102150405")
	}

	// 获取AI的响应
	aiResponse, err := dao.GetAIResponse(models.AIRequest{
		Message:   req.Message,
		SessionID: sessionID,
	})
	if err != nil {
		log.Printf("AI ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回AI的响应
	c.JSON(http.StatusOK, aiResponse)
}
