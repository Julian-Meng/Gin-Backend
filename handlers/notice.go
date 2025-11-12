package handlers

import (
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// =============================
// 获取公告列表（所有用户可访问）
// =============================
func GetAllNotices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	notices, total, err := models.GetAllNotices(page, pageSize)
	if err != nil {
		log.Println("❌ 获取公告失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "获取公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "success",
		"total": total,
		"data":  notices,
	})
}

// =============================
// 获取单条公告详情
// =============================
func GetNoticeByID(c *gin.Context) {
	id := c.Param("id")

	notice, err := models.GetNoticeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
}

// =============================
// 创建公告（仅管理员）
// =============================
func CreateNotice(c *gin.Context) {
	var n models.Notice
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	username, _ := c.Get("username")
	if username == nil {
		username = "未知管理员"
	}
	n.Publisher = username.(string)

	if err := models.CreateNotice(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告创建成功"})
}

// =============================
// 更新公告
// =============================
func UpdateNotice(c *gin.Context) {
	id := c.Param("id")
	var n models.Notice
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请求格式错误"})
		return
	}

	if err := models.UpdateNotice(id, n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告已更新"})
}

// =============================
// 删除公告
// =============================
func DeleteNotice(c *gin.Context) {
	id := c.Param("id")

	if err := models.DeleteNotice(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除公告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告已删除"})
}
