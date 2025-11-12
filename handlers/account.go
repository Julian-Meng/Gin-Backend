package handlers

import (
	"backend/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取全部账号（管理员权限）
func GetAllAccounts(c *gin.Context) {
	accounts := models.GetAllAccounts()
	c.JSON(http.StatusOK, gin.H{"data": accounts})
}

// 更新账号角色或状态
func UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var acc models.Account
	if err := c.BindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := models.UpdateAccount(id, acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("更新失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号更新成功"})
}

// 删除账号
func DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if err := models.DeleteAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "账号删除成功"})
}
