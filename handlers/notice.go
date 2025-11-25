package handlers

import (
	"backend/dao"
	"backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ==========================
// 获取公告列表（分页）
// GET /notices?page=&pageSize=
// ==========================
func GetAllNotices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	list, total, err := dao.GetAllNotices(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "获取公告列表失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// ==========================
// 获取单条公告
// GET /notices/:id
// ==========================
func GetNoticeByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的公告 ID",
		})
		return
	}

	n, err := dao.GetNoticeByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "公告不存在",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": n,
	})
}

// ==========================
// 创建公告
// POST /notices
// Body: { "title": "...", "content": "...", "publisher": "..." }
// ==========================
func CreateNotice(c *gin.Context) {
	var req models.Notice

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
			"err":  err.Error(),
		})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Publisher = strings.TrimSpace(req.Publisher)

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "公告标题不能为空",
		})
		return
	}

	if err := dao.CreateNotice(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "创建公告失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "公告创建成功",
	})
}

// ==========================
// 更新公告
// PUT /notices/:id
// ==========================
func UpdateNotice(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的公告 ID",
		})
		return
	}

	var req models.Notice
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
			"err":  err.Error(),
		})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Publisher = strings.TrimSpace(req.Publisher)

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "公告标题不能为空",
		})
		return
	}

	if err := dao.UpdateNotice(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "公告更新失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告更新成功"})
}

// ==========================
// 删除公告
// DELETE /notices/:id
// ==========================
func DeleteNotice(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的公告 ID",
		})
		return
	}

	if err := dao.DeleteNotice(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除公告失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告删除成功"})
}
