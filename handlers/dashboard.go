package handlers

import (
	"backend/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 管理员仪表盘
// GET /dashboard/admin
func AdminDashboard(c *gin.Context) {
	data := dao.GetAdminDashboardData()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

// 普通用户仪表盘
// GET /dashboard/user
func UserDashboard(c *gin.Context) {
	data := dao.GetUserDashboardData()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}
