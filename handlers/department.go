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

// GetDepartments godoc
// @Summary 管理员分页获取部门列表
// @Tags department
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "关键字"
// @Success 200 {object} DepartmentListResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/departments [get]
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
		errorx.Internal(c, "获取部门列表失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// GetDepartmentByID godoc
// @Summary 获取部门详情
// @Tags department
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} DepartmentDetailResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/admin/department/{id} [get]
// @Router /api/user/department/{id} [get]
func GetDepartmentByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的部门 ID", err)
		return
	}

	d, err := dao.FetchDepartmentByID(uint(id64))
	if err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "部门不存在", err)
			return
		}
		errorx.Internal(c, "查询部门失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": d,
	})
}

// CreateDepartment godoc
// @Summary 创建部门
// @Tags department
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Department true "部门信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/department [post]
func CreateDepartment(c *gin.Context) {
	var req models.Department
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	// 基础校验
	if req.Name == "" {
		errorx.BadRequest(c, "部门名称不能为空", nil)
		return
	}
	if req.FullNum < 1 {
		req.FullNum = 20
	}

	// 创建
	if err := dao.InsertDepartment(req); err != nil {
		errorx.Internal(c, "创建部门失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "部门创建成功",
	})
}

// UpdateDepartment godoc
// @Summary 更新部门
// @Tags department
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param request body models.Department true "部门信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/department/{id} [put]
func UpdateDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的部门 ID", err)
		return
	}

	var req models.Department
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		errorx.BadRequest(c, "部门名称不能为空", nil)
		return
	}

	if req.FullNum < 1 {
		req.FullNum = 20
	}
	if req.DptNum < 0 {
		req.DptNum = 0
	}

	if err := dao.UpdateDepartment(uint(id64), req); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "部门不存在", err)
			return
		}
		errorx.Internal(c, "更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门更新成功"})
}

// DeleteDepartment godoc
// @Summary 删除部门
// @Tags department
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Router /api/admin/department/{id} [delete]
func DeleteDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的部门 ID", err)
		return
	}

	err = dao.DeleteDepartment(uint(id64))
	if err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "部门不存在", err)
			return
		}
		// 部门内有人时给出明确提示
		errorx.BadRequest(c, err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门删除成功"})
}
