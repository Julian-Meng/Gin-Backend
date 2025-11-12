package handlers

import (
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 管理员仪表盘
func AdminDashboard(c *gin.Context) {
	data := models.GetAdminDashboardData()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

// 普通用户仪表盘
func UserDashboard(c *gin.Context) {
	data := models.GetUserDashboardData()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}
