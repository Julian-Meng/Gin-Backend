package handlers

import (
	"backend/dao"
	"backend/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
//	   "target_dpt": 3,
//	   "change_type": 1/2/3/4,
//	   "description": "...",
//	   "leave_start_at": "2026-04-10", // change_type=4 必填（支持 YYYY-MM-DD / RFC3339）
//	   "leave_end_at": "2026-04-11",   // change_type=4 必填（支持 YYYY-MM-DD / RFC3339）
//	   "leave_reason": "..."
//	}
func CreatePersonnel(c *gin.Context) {
	var req struct {
		TargetDpt    uint   `json:"target_dpt"`
		ChangeType   int    `json:"change_type"`
		Description  string `json:"description"`
		LeaveStartAt string `json:"leave_start_at"`
		LeaveEndAt   string `json:"leave_end_at"`
		LeaveReason  string `json:"leave_reason"`
		LeaveType    string `json:"leave_type"`
		HandoverNote string `json:"handover_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
		})
		return
	}

	req.Description = strings.TrimSpace(req.Description)
	req.LeaveStartAt = strings.TrimSpace(req.LeaveStartAt)
	req.LeaveEndAt = strings.TrimSpace(req.LeaveEndAt)
	req.LeaveReason = strings.TrimSpace(req.LeaveReason)
	req.LeaveType = strings.TrimSpace(req.LeaveType)
	req.HandoverNote = strings.TrimSpace(req.HandoverNote)

	empVal, exists := c.Get("emp_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未获取到登录用户身份信息"})
		return
	}
	empID, _ := empVal.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "当前登录账号未绑定员工工号，无法提交申请"})
		return
	}
	if req.ChangeType < 1 || req.ChangeType > 4 {
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
	var leaveStartAt *time.Time
	var leaveEndAt *time.Time
	if req.ChangeType == 4 {
		if req.LeaveStartAt == "" || req.LeaveEndAt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请假必须提供 leave_start_at 和 leave_end_at"})
			return
		}

		start, err := parseLeaveDate(req.LeaveStartAt, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "leave_start_at 格式错误，应为 YYYY-MM-DD 或 RFC3339"})
			return
		}
		end, err := parseLeaveDate(req.LeaveEndAt, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "leave_end_at 格式错误，应为 YYYY-MM-DD 或 RFC3339"})
			return
		}
		if end.Before(start) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请假结束日期不能早于开始日期"})
			return
		}
		if req.LeaveReason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请假必须提供 leave_reason"})
			return
		}

		leaveStartAt = &start
		leaveEndAt = &end
	}

	record := models.Personnel{
		EmpID:        empID,
		TargetDpt:    req.TargetDpt,
		ChangeType:   req.ChangeType,
		Description:  req.Description,
		LeaveStartAt: leaveStartAt,
		LeaveEndAt:   leaveEndAt,
		LeaveReason:  req.LeaveReason,
		LeaveType:    req.LeaveType,
		HandoverNote: req.HandoverNote,
		State:        0, // 待审批
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

func parseLeaveDate(raw string, endOfDay bool) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	if t, err := time.Parse("2006-01-02", raw); err == nil {
		if endOfDay {
			return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local), nil
		}
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
	}

	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		lt := t.Local()
		if endOfDay {
			return time.Date(lt.Year(), lt.Month(), lt.Day(), 23, 59, 59, 0, time.Local), nil
		}
		return time.Date(lt.Year(), lt.Month(), lt.Day(), 0, 0, 0, 0, time.Local), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

// ApprovePersonnel 审批人事变更
// PUT /personnel/approve
//
//	Body {
//	   "id": 123,
//	   "approver": "admin_name",
//	   "approve": true/false,
//	   "reject_reason": "..." // 驳回时可填
//	}
func ApprovePersonnel(c *gin.Context) {
	var req struct {
		ID           uint   `json:"id"`
		Approver     string `json:"approver"`
		Approve      bool   `json:"approve"`
		RejectReason string `json:"reject_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "参数解析失败",
		})
		return
	}

	req.Approver = strings.TrimSpace(req.Approver)
	req.RejectReason = strings.TrimSpace(req.RejectReason)
	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "ID 不能为空"})
		return
	}
	if req.Approver == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "审批人不能为空"})
		return
	}
	if !req.Approve && req.RejectReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "驳回时必须提供 reject_reason"})
		return
	}

	// 审批逻辑（部门人数增减 / 调岗 / 离职 全部由 DAO 处理）
	if err := dao.ApprovePersonnelChange(req.ID, req.Approver, req.Approve, req.RejectReason); err != nil {
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
