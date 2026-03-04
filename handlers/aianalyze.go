package handlers

import (
	"backend/dao"
	"backend/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAnalyzeDashboardAI 管理员：AI 解读仪表盘（全局数据）
// GET /api/admin/ai/analyze/dashboard
func AdminAnalyzeDashboardAI(c *gin.Context) {
	data := dao.GetAdminDashboardData()

	analysis, err := analyzeDashboardWithAI(data, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "failed to analyze dashboard by AI",
			"data": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"analysis": analysis},
	})
}

// UserAnalyzeDashboardAI 普通用户：AI 解读仪表盘（个人数据）
// GET /api/user/ai/analyze/dashboard
func UserAnalyzeDashboardAI(c *gin.Context) {
	data := dao.GetUserDashboardData()

	analysis, err := analyzeDashboardWithAI(data, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "failed to analyze dashboard by AI",
			"data": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"analysis": analysis},
	})
}

// 内部函数：把 dashboard 数据喂给 AI，让 AI 输出 “概览/潜在问题/建议” 三段式解读
func analyzeDashboardWithAI(dashboard interface{}, isAdmin bool) (string, error) {
	raw, err := json.Marshal(dashboard)
	if err != nil {
		return "", err
	}

	roleHint := "这是普通员工的个人仪表盘数据。"
	if isAdmin {
		roleHint = "这是管理员看到的全局仪表盘数据。"
	}

	prompt := roleHint + `
请你根据以下 JSON 数据，生成一段简洁的业务解读报告，要求：
1）用 3 个小标题输出：【概览】【潜在问题】【建议】
2）每个小标题下用 2~4 条要点，避免长段落，不要超过 150 字
3）只输出最终结果，不要输出任何推理过程/思考过程/think 标签
4）如果数据不足以判断，就明确说“暂无足够数据判断”，不要编造

仪表盘 JSON 数据如下：
` + string(raw)

	resp, err := dao.GetAIResponse(models.AIRequest{
		Message:   prompt,
		SessionID: "dashboard-analysis",
	})
	if err != nil {
		return "", err
	}

	return resp.Message, nil
}
