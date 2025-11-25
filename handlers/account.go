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
// 获取全部账号（建议管理员权限）
// GET /accounts
// ==========================
func GetAllAccounts(c *gin.Context) {
	accounts := dao.GetAllAccounts()

	// 安全起见，不把密码返回给前端（即便是 hash）
	for i := range accounts {
		accounts[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": accounts,
	})
}

// ==========================
// 创建账号（注册 / 后台新增）
// POST /accounts
// Body: { "username": "...", "password": "...", "role": "admin|staff" }
// ==========================
func CreateAccount(c *gin.Context) {
	var req models.Account

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求体解析失败",
			"err":  err.Error(),
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名不能为空",
		})
		return
	}

	// 简单长度检查，可按需调整
	if len(req.Username) < 3 || len(req.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名长度需在 3-32 之间",
		})
		return
	}

	// 检查重名
	if _, exists := dao.GetAccountByUsername(req.Username); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名已存在",
		})
		return
	}

	// 角色兜底
	if req.Role == "" {
		req.Role = "staff"
	}

	// 调用 DAO，内部会做：
	// - 默认密码（为空则 123456）
	// - 密码加密
	// - 生成 EmpID
	// - 创建 PERSON 记录
	if err := dao.InsertAccount(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "创建账号失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号创建成功",
	})
}

// ==========================
// 更新账号信息（角色 / 状态）
// PUT /accounts/:id
// Body: { "role": "admin|staff", "status": 0|1 }
// ==========================
func UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的账号 ID",
		})
		return
	}

	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求体解析失败",
			"err":  err.Error(),
		})
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role != "admin" && req.Role != "staff" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "角色只能为 admin 或 staff",
		})
		return
	}

	// 状态：允许 0 / 1，其他当非法
	if req.Status != 0 && req.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "状态只能为 0 或 1",
		})
		return
	}

	if err := dao.UpdateAccount(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "账号更新失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号更新成功",
	})
}

// ==========================
// 删除账号
// DELETE /accounts/:id
// ==========================
func DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的账号 ID",
		})
		return
	}

	if err := dao.DeleteAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除账号失败",
			"err":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号删除成功",
	})
}
