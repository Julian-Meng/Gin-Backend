package handlers

import (
	"backend/dao"
	"backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetPersonnelList 获取人事变更列表（分页）
// GET /personnel?page=&pageSize=
func GetPersonnelList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	list, total, err := dao.FetchPersonnelPaged(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "获取人事变更记录失败",
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

// GetPersonnelByID 获取单条人事变更详情
// GET /personnel/:id
func GetPersonnelByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的 ID",
		})
		return
	}

	data, err := dao.GetPersonnelByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "记录不存在",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": data,
	})
}

// CreatePersonnel 发起人事变更申请
// POST /personnel
//
//	Body {
//	   "emp_id": "...",
//	   "target_dpt": 3,
//	   "change_type": 1/2/3,
//	   "description": "..."
//	}
func CreatePersonnel(c *gin.Context) {
	var req struct {
		EmpID       string `json:"emp_id"`
		TargetDpt   uint   `json:"target_dpt"`
		ChangeType  int    `json:"change_type"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
		})
		return
	}

	req.EmpID = strings.TrimSpace(req.EmpID)
	req.Description = strings.TrimSpace(req.Description)

	if req.EmpID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "emp_id 不能为空"})
		return
	}
	if req.ChangeType < 1 || req.ChangeType > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的变更类型"})
		return
	}
	if req.ChangeType == 1 && req.TargetDpt == 0 { // 调部门必须指定目标部门
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "调部门必须提供 target_dpt"})
		return
	}
	if req.ChangeType == 2 && req.Description == "" { // 调岗必须提供岗位名
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "调岗必须提供 description（新岗位名）"})
		return
	}

	record := models.Personnel{
		EmpID:       req.EmpID,
		TargetDpt:   req.TargetDpt,
		ChangeType:  req.ChangeType,
		Description: req.Description,
		State:       0, // 待审批
	}

	if err := dao.CreatePersonnelChange(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "提交人事变更失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "申请已提交"})
}

// ApprovePersonnel 审批人事变更
// PUT /personnel/approve
//
//	Body {
//	   "id": 123,
//	   "approver": "admin_name",
//	   "approve": true/false
//	}
func ApprovePersonnel(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id"`
		Approver string `json:"approver"`
		Approve  bool   `json:"approve"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
		})
		return
	}

	req.Approver = strings.TrimSpace(req.Approver)
	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "ID 不能为空"})
		return
	}
	if req.Approver == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "审批人不能为空"})
		return
	}

	// 审批逻辑（部门人数增减 / 调岗 / 离职 全部由 DAO 处理）
	if err := dao.ApprovePersonnelChange(req.ID, req.Approver, req.Approve); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "审批失败",
			"err":  err.Error(),
		})
		return
	}

	msg := "已驳回"
	if req.Approve {
		msg = "审批通过"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  msg,
	})
}
