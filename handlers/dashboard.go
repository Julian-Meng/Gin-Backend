package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminDashboard godoc
// @Summary 管理员仪表盘
// @Description 返回全局统计、待处理变更与近期公告
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardAdminResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 403 {object} APIErrorResponse
// @Router /api/admin/dashboard [get]
func AdminDashboard(c *gin.Context) {
	data, err := dao.GetAdminDashboardData()
	if err != nil {
		errorx.Internal(c, "获取管理员仪表盘失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

// UserDashboard godoc
// @Summary 用户仪表盘
// @Description 返回个人相关统计与近期公告
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardUserResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 403 {object} APIErrorResponse
// @Router /api/user/dashboard [get]
func UserDashboard(c *gin.Context) {
	data, err := dao.GetUserDashboardData()
	if err != nil {
		errorx.Internal(c, "获取用户仪表盘失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}
