package handlers

import (
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ✅ 获取部门列表（支持分页与搜索）
func GetAllDepartments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.DefaultQuery("keyword", "")

	list, total, err := models.FetchDepartmentsPaged(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "查询部门失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"list":  list,
			"total": total,
		},
	})

}

// ✅ 按 ID 查询部门
func GetDepartmentByID(c *gin.Context) {
	id := c.Param("id")
	dpt, err := models.FetchDepartmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该部门"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dpt})
}

// ✅ 新增部门
func CreateDepartment(c *gin.Context) {
	var d models.Department
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := models.InsertDepartment(d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "新增成功"})
}

// ✅ 更新部门信息
func UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	var d models.Department
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := models.UpdateDepartment(id, d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// ✅ 删除部门
func DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	if err := models.DeleteDepartment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
