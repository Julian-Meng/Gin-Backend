package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"backend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPersons godoc
// @Summary 管理员分页获取员工列表
// @Tags person
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "关键字"
// @Success 200 {object} PersonListResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/persons [get]
func GetPersons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := strings.TrimSpace(c.DefaultQuery("keyword", ""))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	list, total, err := dao.FetchPersonsPaged(page, pageSize, keyword)
	if err != nil {
		errorx.Internal(c, "获取员工列表失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// GetPersonByID godoc
// @Summary 获取员工详情
// @Tags person
// @Produce json
// @Security BearerAuth
// @Param id path int true "员工ID"
// @Success 200 {object} PersonDetailResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/admin/person/{id} [get]
// @Router /api/user/profile/{id} [get]
func GetPersonByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的员工 ID", err)
		return
	}

	p, err := dao.FetchPersonByID(uint(id64))
	if err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "查询员工失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": p,
	})
}

// CreatePerson godoc
// @Summary 创建员工
// @Tags person
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Person true "员工信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person [post]
func CreatePerson(c *gin.Context) {
	var req models.Person
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "请求解析失败", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		errorx.BadRequest(c, "姓名不能为空", nil)
		return
	}

	// Birth 转换
	if req.Birth != nil {
		if req.Birth.After(time.Now()) {
			errorx.BadRequest(c, "出生日期不能晚于当前时间", nil)
			return
		}
	}

	if err := dao.CreatePerson(&req); err != nil {
		errorx.Internal(c, "创建员工失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工创建成功"})
}

// UpdatePerson godoc
// @Summary 更新员工信息
// @Description 更新基础资料（不含专门的岗位/状态/部门调整接口）
// @Tags person
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "员工ID"
// @Param request body models.Person true "员工信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/{id} [put]
// @Router /api/user/profile/{id} [put]
func UpdatePerson(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的员工 ID", err)
		return
	}

	var req models.Person
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "请求解析失败", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		errorx.BadRequest(c, "姓名不能为空", nil)
		return
	}

	if req.Birth != nil && req.Birth.After(time.Now()) {
		errorx.BadRequest(c, "出生日期不能晚于当前时间", nil)
		return
	}

	if err := dao.UpdatePerson(uint(id64), req); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工信息更新成功"})
}

// DeletePersonByEmpID godoc
// @Summary 按员工编号删除员工
// @Tags person
// @Produce json
// @Security BearerAuth
// @Param emp_id path string true "员工编号"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/emp/{emp_id} [delete]
func DeletePersonByEmpID(c *gin.Context) {
	empID := strings.TrimSpace(c.Param("emp_id"))
	if empID == "" {
		errorx.BadRequest(c, "无效的员工编号", nil)
		return
	}

	if err := dao.DeletePersonByEmpID(empID); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "删除失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工已删除"})
}

// DeletePersonByID godoc
// @Summary 按ID删除员工
// @Tags person
// @Produce json
// @Security BearerAuth
// @Param id path int true "员工ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/{id} [delete]
func DeletePersonByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的员工 ID", err)
		return
	}

	if err := dao.SafeDeletePerson(uint(id64)); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "删除失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工已删除"})
}

// ChangePersonDepartment godoc
// @Summary 调整员工部门
// @Tags person
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PersonDepartmentChangeRequest true "部门调整参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/change-dept [put]
func ChangePersonDepartment(c *gin.Context) {
	var req PersonDepartmentChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	req.EmpID = strings.TrimSpace(req.EmpID)
	req.Dept = strings.TrimSpace(req.Dept)

	if req.EmpID == "" || req.Dept == "" {
		errorx.BadRequest(c, "emp_id 和 dept 均不能为空", nil)
		return
	}

	if err := dao.UpdatePersonDepartment(req.EmpID, req.Dept); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工或部门不存在", err)
			return
		}
		errorx.Internal(c, "部门调整失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门调整成功"})
}

// ChangePersonState godoc
// @Summary 修改员工状态
// @Tags person
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PersonStateChangeRequest true "状态调整参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/state [put]
func ChangePersonState(c *gin.Context) {
	var req PersonStateChangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	if req.EmpID == "" || (req.State != 0 && req.State != 1) {
		errorx.BadRequest(c, "参数不合法", nil)
		return
	}

	if err := dao.UpdatePersonState(req.EmpID, req.State); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "状态更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态更新成功"})
}

// ChangePersonJob godoc
// @Summary 修改员工岗位
// @Tags person
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PersonJobChangeRequest true "岗位调整参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/person/job [put]
func ChangePersonJob(c *gin.Context) {
	var req PersonJobChangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	if req.EmpID == "" || strings.TrimSpace(req.Job) == "" {
		errorx.BadRequest(c, "emp_id 与 job 均不能为空", nil)
		return
	}

	if err := dao.UpdatePersonJob(req.EmpID, req.Job); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "员工不存在", err)
			return
		}
		errorx.Internal(c, "职位更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "职位更新成功"})
}
