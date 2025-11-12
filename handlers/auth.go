package handlers

import (
	"backend/middlewares"
	"backend/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==========================
// 登录逻辑
// ==========================
func Login(c *gin.Context) {
	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  1,
			"msg":   "请求格式错误",
			"error": err.Error(),
		})
		return
	}

	// ✅ root 超级管理员特权（无密码也可登录）
	if req.Username == "root" {
		token, err := middlewares.GenerateToken("root", "superadmin")
		if err != nil {
			log.Println("❌ 生成 root Token 失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "生成 Token 失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":     0,
			"msg":      "登录成功",
			"token":    token,
			"username": "root",
			"role":     "superadmin",
		})
		return
	}

	// ✅ 普通用户登录验证
	account, ok := models.ValidateLogin(req.Username, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "用户名或密码错误",
		})
		return
	}

	// 生成 Token
	token, err := middlewares.GenerateToken(account.Username, account.Role)
	if err != nil {
		log.Println("❌ 生成 Token 失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "Token 生成失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"msg":      "登录成功",
		"token":    token,
		"username": account.Username,
		"role":     account.Role,
	})
}

// ==========================
// 注册逻辑
// ==========================
func Register(c *gin.Context) {
	var acc models.Account
	if err := c.ShouldBindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  1,
			"msg":   "请求格式错误",
			"error": err.Error(),
		})
		return
	}

	// ✅ 防空处理：若密码为空，则设置默认密码
	if acc.Password == "" {
		acc.Password = "123456"
	}

	// ✅ 插入数据库
	if err := models.InsertAccount(acc); err != nil {
		log.Println("❌ 注册失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "注册失败",
			"error": fmt.Sprintf("%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": true,
		"msg":     "注册成功",
	})
}
