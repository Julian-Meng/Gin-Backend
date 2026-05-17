package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAllAccounts godoc
// @Summary 获取账号列表
// @Tags account
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AccountListResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 403 {object} APIErrorResponse
// @Router /api/admin/accounts [get]
func GetAllAccounts(c *gin.Context) {
	accounts, err := dao.GetAllAccounts()
	if err != nil {
		errorx.Internal(c, "获取账号列表失败", err)
		return
	}

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

// CreateAccount godoc
// @Summary 创建账号
// @Tags account
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AccountCreateRequest true "账号信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/account [post]
func CreateAccount(c *gin.Context) {
	var req models.Account

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "请求体解析失败", err)
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" {
		errorx.BadRequest(c, "用户名不能为空", nil)
		return
	}

	// 简单长度检查，可按需调整
	if len(req.Username) < 3 || len(req.Username) > 32 {
		errorx.BadRequest(c, "用户名长度需在 3-32 之间", nil)
		return
	}

	// 检查重名
	if _, err := dao.GetAccountByUsername(req.Username); err == nil {
		errorx.Conflict(c, "用户名已存在", nil)
		return
	} else if !dao.IsNotFound(err) {
		errorx.Internal(c, "校验用户名是否存在失败", err)
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
		errorx.Internal(c, "创建账号失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号创建成功",
	})
}

// UpdateAccount godoc
// @Summary 更新账号信息
// @Description 更新角色与状态字段
// @Tags account
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "账号ID"
// @Param request body AccountUpdateRequest true "更新参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/account/{id} [put]
func UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		errorx.BadRequest(c, "无效的账号 ID", err)
		return
	}

	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "请求体解析失败", err)
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role != "admin" && req.Role != "staff" {
		errorx.BadRequest(c, "角色只能为 admin 或 staff", nil)
		return
	}

	// 状态：允许 0 / 1，其他当非法
	if req.Status != 0 && req.Status != 1 {
		errorx.BadRequest(c, "状态只能为 0 或 1", nil)
		return
	}

	if err := dao.UpdateAccount(id, req); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "账号不存在", err)
			return
		}
		errorx.Internal(c, "账号更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号更新成功",
	})
}

// DeleteAccount godoc
// @Summary 删除账号
// @Tags account
// @Produce json
// @Security BearerAuth
// @Param id path int true "账号ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/account/{id} [delete]
func DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		errorx.BadRequest(c, "无效的账号 ID", err)
		return
	}

	if err := dao.DeleteAccount(id); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "账号不存在", err)
			return
		}
		errorx.Internal(c, "删除账号失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "账号删除成功",
	})
}
