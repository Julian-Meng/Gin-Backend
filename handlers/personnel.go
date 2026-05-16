package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"backend/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPersonnelList godoc
// @Summary 管理员分页获取人事变更列表
// @Tags personnel
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} PersonnelListResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/changes [get]
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

// GetMyPersonnelList godoc
// @Summary 获取当前用户的人事变更列表
// @Tags personnel
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param page_size query int false "每页数量(兼容字段)"
// @Success 200 {object} PersonnelListResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/user/changes [get]
func GetMyPersonnelList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	pageSizeRaw := strings.TrimSpace(c.Query("pageSize"))
	if pageSizeRaw == "" {
		pageSizeRaw = strings.TrimSpace(c.Query("page_size"))
	}
	if pageSizeRaw == "" {
		pageSizeRaw = "10"
	}
	pageSize, _ := strconv.Atoi(pageSizeRaw)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	empVal, exists := c.Get("emp_id")
	if !exists {
		errorx.Unauthorized(c, "未获取到登录用户身份信息", nil)
		return
	}

	empID, _ := empVal.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		errorx.BadRequest(c, "当前登录账号未绑定员工工号，无法获取人事变更记录", nil)
		return
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	list, total, err := dao.FetchPersonnelByEmpIDPaged(empID, limit, offset)
	if err != nil {
		errorx.Internal(c, "获取我的人事变更记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// GetPersonnelByID godoc
// @Summary 获取人事变更详情
// @Tags personnel
// @Produce json
// @Security BearerAuth
// @Param id path int true "变更记录ID"
// @Success 200 {object} PersonnelDetailResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/admin/change/{id} [get]
func GetPersonnelByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的 ID", fmt.Errorf("invalid id: %s", idStr))
		return
	}

	data, err := dao.GetPersonnelByID(uint(id64))
	if err != nil {
		errorx.NotFound(c, "记录不存在", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": data,
	})
}

// CreatePersonnel godoc
// @Summary 发起人事变更申请
// @Description 管理员与普通用户都可调用，申请人 emp_id 从 JWT 获取
// @Tags personnel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PersonnelCreateRequest true "变更申请参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/change [post]
// @Router /api/user/change/request [post]
func CreatePersonnel(c *gin.Context) {
	var req PersonnelCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
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
		errorx.Unauthorized(c, "未获取到登录用户身份信息", nil)
		return
	}
	empID, _ := empVal.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		errorx.BadRequest(c, "当前登录账号未绑定员工工号，无法提交申请", nil)
		return
	}
	if req.ChangeType < 1 || req.ChangeType > 4 {
		errorx.BadRequest(c, "无效的变更类型", fmt.Errorf("invalid change_type: %d", req.ChangeType))
		return
	}
	if req.ChangeType == 1 && req.TargetDpt == 0 { // 调部门必须指定目标部门
		errorx.BadRequest(c, "调部门必须提供 target_dpt", nil)
		return
	}
	if req.ChangeType == 2 && req.Description == "" { // 调岗必须提供岗位名
		errorx.BadRequest(c, "调岗必须提供 description（新岗位名）", nil)
		return
	}
	var leaveStartAt *time.Time
	var leaveEndAt *time.Time
	if req.ChangeType == 4 {
		if req.LeaveStartAt == "" || req.LeaveEndAt == "" {
			errorx.BadRequest(c, "请假必须提供 leave_start_at 和 leave_end_at", nil)
			return
		}

		start, err := parseLeaveDate(req.LeaveStartAt, false)
		if err != nil {
			errorx.BadRequest(c, "leave_start_at 格式错误，应为 YYYY-MM-DD 或 RFC3339", nil)
			return
		}
		end, err := parseLeaveDate(req.LeaveEndAt, true)
		if err != nil {
			errorx.BadRequest(c, "leave_end_at 格式错误，应为 YYYY-MM-DD 或 RFC3339", nil)
			return
		}
		if end.Before(start) {
			errorx.BadRequest(c, "请假结束日期不能早于开始日期", nil)
			return
		}
		if req.LeaveReason == "" {
			errorx.BadRequest(c, "请假必须提供 leave_reason", nil)
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
		errorx.Internal(c, "提交人事变更失败", err)
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

// ApprovePersonnel godoc
// @Summary 审批人事变更
// @Tags personnel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PersonnelApproveRequest true "审批参数"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 401 {object} APIErrorResponse
// @Failure 403 {object} APIErrorResponse
// @Router /api/admin/change/approve [put]
func ApprovePersonnel(c *gin.Context) {
	var req PersonnelApproveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", nil)
		return
	}

	req.Approver = strings.TrimSpace(req.Approver)
	req.RejectReason = strings.TrimSpace(req.RejectReason)
	if req.ID == 0 {
		errorx.BadRequest(c, "ID 不能为空", nil)
		return
	}
	if req.Approver == "" {
		errorx.BadRequest(c, "审批人不能为空", nil)
		return
	}
	if !req.Approve && req.RejectReason == "" {
		errorx.BadRequest(c, "驳回时必须提供 reject_reason", nil)
		return
	}

	// 审批逻辑（部门人数增减 / 调岗 / 离职 全部由 DAO 处理）
	if err := dao.ApprovePersonnelChange(req.ID, req.Approver, req.Approve, req.RejectReason); err != nil {
		errorx.BadRequest(c, "审批失败", err)
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
