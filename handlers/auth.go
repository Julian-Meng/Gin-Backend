package handlers

import (
	"backend/dao"
	"backend/middlewares"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ==========================
// superadmin / root 配置
// ==========================
// 这里先用硬编码，建议后面改成从 env / config 读取
const superAdminUsername = "root"
const superAdminPassword = "root123"
const superAdminRole = "superadmin"

// ==========================
// 登录逻辑
// POST /auth/login
// Body: { "username": "...", "password": "..." }
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

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名和密码不能为空",
		})
		return
	}

	// ==========================
	// 1. root 超级管理员特权
	// ==========================
	if req.Username == superAdminUsername {
		if req.Password != superAdminPassword {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "root 密码错误",
			})
			return
		}
		// root 超级管理员
		token, err := middlewares.GenerateToken(superAdminUsername, "", superAdminRole)
		if err != nil {
			log.Println("❌ 生成 root Token 失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "生成 Token 失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":     0,
			"msg":      "登录成功",
			"token":    token,
			"username": superAdminUsername,
			"role":     superAdminRole,
		})
		return
	}

	// ==========================
	// 2. 普通用户登录（走 DAO）
	// ==========================
	account, ok := dao.ValidateLogin(req.Username, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "用户名或密码错误",
		})
		return
	}

	//普通用户（有账号 + emp_id）
	token, err := middlewares.GenerateToken(account.Username, account.EmpID, account.Role)
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
// POST /auth/register
// Body: { "username": "...", "password": "...", "role": "staff|admin" }
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

	acc.Username = strings.TrimSpace(acc.Username)

	if acc.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名不能为空",
		})
		return
	}

	// 禁止通过注册接口创建 / 覆盖 root 账号
	if acc.Username == superAdminUsername {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 1,
			"msg":  "不允许注册保留用户名",
		})
		return
	}

	// 可选：用户名长度限制
	if len(acc.Username) < 3 || len(acc.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名长度需在 3-32 之间",
		})
		return
	}

	// 若密码为空，交给 DAO 里兜底（默认为 123456），也可以这里兜一层
	if strings.TrimSpace(acc.Password) == "" {
		acc.Password = ""
	}

	// 检查重名（原来是直接 InsertAccount，遇到唯一索引冲突就 500）:contentReference[oaicite:2]{index=2}
	if _, exists := dao.GetAccountByUsername(acc.Username); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名已存在",
		})
		return
	}

	// 调用 DAO：
	// - 默认密码处理
	// - 密码加密
	// - 生成 EmpID
	// - 创建 PERSON 记录
	if err := dao.InsertAccount(acc); err != nil {
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
