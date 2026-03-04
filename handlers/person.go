package handlers

import (
	"backend/dao"
	"backend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPersons 分页获取员工列表
// GET /persons?page=&pageSize=&keyword=
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "获取员工列表失败",
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

// GetPersonByID 获取员工详情（ID）
// GET /persons/:id
func GetPersonByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的员工 ID"})
		return
	}

	p, err := dao.FetchPersonByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "员工不存在",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": p,
	})
}

// CreatePerson 创建员工
// POST /persons
func CreatePerson(c *gin.Context) {
	var req models.Person
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请求解析失败", "err": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "姓名不能为空"})
		return
	}

	// Birth 转换
	if req.Birth != nil {
		if req.Birth.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "出生日期不能晚于当前时间"})
			return
		}
	}

	if err := dao.CreatePerson(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建员工失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工创建成功"})
}

// UpdatePerson 更新员工信息（不含部门、职位、离职）
// PUT /persons/:id
func UpdatePerson(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的员工 ID"})
		return
	}

	var req models.Person
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请求解析失败", "err": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "姓名不能为空"})
		return
	}

	if req.Birth != nil && req.Birth.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "出生日期不能晚于当前时间"})
		return
	}

	if err := dao.UpdatePerson(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工信息更新成功"})
}

// DeletePersonByEmpID 删除员工（依 emp_id）
// DELETE /persons/emp/:emp_id
func DeletePersonByEmpID(c *gin.Context) {
	empID := strings.TrimSpace(c.Param("emp_id"))
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的员工编号"})
		return
	}

	if err := dao.DeletePersonByEmpID(empID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工已删除"})
}

// DeletePersonByID 删除员工（按 ID）
// DELETE /persons/:id
func DeletePersonByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的员工 ID"})
		return
	}

	if err := dao.SafeDeletePerson(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "员工已删除"})
}

// ChangePersonDepartment 更改员工部门
// PUT /persons/change-dept
// Body: { "emp_id": "...", "dept": "部门名称" }
func ChangePersonDepartment(c *gin.Context) {
	var req struct {
		EmpID string `json:"emp_id"`
		Dept  string `json:"dept"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数解析失败"})
		return
	}

	req.EmpID = strings.TrimSpace(req.EmpID)
	req.Dept = strings.TrimSpace(req.Dept)

	if req.EmpID == "" || req.Dept == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "emp_id 和 dept 均不能为空"})
		return
	}

	if err := dao.UpdatePersonDepartment(req.EmpID, req.Dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "部门调整失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "部门调整成功"})
}

// ChangePersonState 修改员工状态（离职或在职）
// PUT /persons/state
// Body: { "emp_id": "...", "state": 0/1 }
func ChangePersonState(c *gin.Context) {
	var req struct {
		EmpID string `json:"emp_id"`
		State int    `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数解析失败"})
		return
	}

	if req.EmpID == "" || (req.State != 0 && req.State != 1) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数不合法"})
		return
	}

	if err := dao.UpdatePersonState(req.EmpID, req.State); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "状态更新失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态更新成功"})
}

// ChangePersonJob 修改员工职位
// PUT /persons/job
// Body: { "emp_id": "...", "job": "..." }
func ChangePersonJob(c *gin.Context) {
	var req struct {
		EmpID string `json:"emp_id"`
		Job   string `json:"job"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数解析失败"})
		return
	}

	if req.EmpID == "" || strings.TrimSpace(req.Job) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "emp_id 与 job 均不能为空"})
		return
	}

	if err := dao.UpdatePersonJob(req.EmpID, req.Job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "职位更新失败", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "职位更新成功"})
}
