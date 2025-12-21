package handlers

import (
	"backend/dao"
	"backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// 获取部门列表（分页 + 搜索）
// GET /departments?page=&pageSize=&keyword=
func GetDepartments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := strings.TrimSpace(c.DefaultQuery("keyword", ""))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	list, total, err := dao.FetchDepartmentsPaged(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "获取部门列表失败",
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

// 获取单个部门（ID）
// GET /departments/:id
func GetDepartmentByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的部门 ID",
		})
		return
	}

	d, err := dao.FetchDepartmentByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "部门不存在",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": d,
	})
}

// 创建部门
// POST /departments
// Body: { "name": "...", "manager": "...", "intro": "...", ... }
func CreateDepartment(c *gin.Context) {
	var req models.Department
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
			"err":  err.Error(),
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	// 基础校验
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "部门名称不能为空",
		})
		return
	}
	if req.FullNum < 1 {
		req.FullNum = 20
	}

	// 创建
	if err := dao.InsertDepartment(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "创建部门失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "部门创建成功",
	})
}

// 更新部门
// PUT /departments/:id
func UpdateDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的部门 ID"})
		return
	}

	var req models.Department
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数解析失败", "err": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "部门名称不能为空"})
		return
	}

	if req.FullNum < 1 {
		req.FullNum = 20
	}
	if req.DptNum < 0 {
		req.DptNum = 0
	}

	if err := dao.UpdateDepartment(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "更新失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门更新成功"})
}

// 删除部门
// DELETE /departments/:id
func DeleteDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的部门 ID"})
		return
	}

	err = dao.DeleteDepartment(uint(id64))
	if err != nil {
		// 部门内有人时给出明确提示
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门删除成功"})
}
