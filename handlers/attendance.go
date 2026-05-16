package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 工具：解析日期（YYYY-MM-DD），带默认值

func parseDateOrDefault(str string, def time.Time) time.Time {
	str = strings.TrimSpace(str)
	if str == "" {
		return def
	}
	// 只解析到日：2006-01-02
	t, err := time.ParseInLocation("2006-01-02", str, time.Local)
	if err != nil {
		return def
	}
	return t
}

// UserCheckIn godoc
// @Summary 用户签到
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Router /api/user/attendance/checkin [post]
func UserCheckIn(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		errorx.BadRequest(c, "当前账号未绑定员工档案，无法签到", nil)
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		errorx.BadRequest(c, "当前账号 emp_id 无效，无法签到", nil)
		return
	}

	if err := dao.CheckIn(empID); err != nil {
		errorx.BadRequest(c, err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "签到成功",
	})
}

// UserCheckOut godoc
// @Summary 用户签退
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Router /api/user/attendance/checkout [post]
func UserCheckOut(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		errorx.BadRequest(c, "当前账号未绑定员工档案，无法签退", nil)
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		errorx.BadRequest(c, "当前账号 emp_id 无效，无法签退", nil)
		return
	}

	if err := dao.CheckOut(empID); err != nil {
		errorx.BadRequest(c, err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "签退成功",
	})
}

// GetMyAttendance godoc
// @Summary 查询我的考勤记录
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Param start query string false "开始日期(YYYY-MM-DD)"
// @Param end query string false "结束日期(YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} AttendanceListResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/user/attendance/my [get]
func GetMyAttendance(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		errorx.BadRequest(c, "当前账号未绑定员工档案，无法查询考勤", nil)
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		errorx.BadRequest(c, "当前账号 emp_id 无效，无法查询考勤", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	// 默认查询最近 30 天
	now := time.Now()
	defaultEnd := now
	defaultStart := now.AddDate(0, 0, -30)

	startStr := c.Query("start")
	endStr := c.Query("end")
	startDate := parseDateOrDefault(startStr, defaultStart)
	endDate := parseDateOrDefault(endStr, defaultEnd)

	list, total, err := dao.GetAttendanceByEmpID(empID, startDate, endDate, page, pageSize)
	if err != nil {
		errorx.Internal(c, "查询考勤记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// AdminSearchAttendance godoc
// @Summary 管理员查询考勤记录
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Param emp_id query string false "员工编号"
// @Param dpt_id query int false "部门ID"
// @Param start query string false "开始日期(YYYY-MM-DD)"
// @Param end query string false "结束日期(YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} AttendanceListResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/attendance [get]
func AdminSearchAttendance(c *gin.Context) {
	empID := strings.TrimSpace(c.Query("emp_id"))
	dptIDStr := strings.TrimSpace(c.Query("dpt_id"))

	var dptID uint64
	var err error
	if dptIDStr != "" {
		dptID, err = strconv.ParseUint(dptIDStr, 10, 64)
		if err != nil {
			errorx.BadRequest(c, "无效的 dpt_id", err)
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	// 默认查询最近 30 天
	now := time.Now()
	defaultEnd := now
	defaultStart := now.AddDate(0, 0, -30)

	startStr := c.Query("start")
	endStr := c.Query("end")
	startDate := parseDateOrDefault(startStr, defaultStart)
	endDate := parseDateOrDefault(endStr, defaultEnd)

	list, total, err := dao.SearchAttendance(empID, uint(dptID), startDate, endDate, page, pageSize)
	if err != nil {
		errorx.Internal(c, "查询考勤记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// AdminUpdateAttendance godoc
// @Summary 管理员更新考勤记录
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "考勤记录ID"
// @Param request body AttendanceUpdateRequest true "更新字段（可部分更新）"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/attendance/{id} [put]
func AdminUpdateAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的考勤记录 ID", err)
		return
	}

	var req AttendanceUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "请求体解析失败", err)
		return
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != nil {
		updates["remark"] = strings.TrimSpace(*req.Remark)
	}
	if req.CheckIn != nil {
		updates["check_in"] = *req.CheckIn
	}
	if req.CheckOut != nil {
		updates["check_out"] = *req.CheckOut
	}

	if len(updates) == 0 {
		errorx.BadRequest(c, "没有任何可更新的字段", nil)
		return
	}

	if err := dao.UpdateAttendance(uint(id64), updates); err != nil {
		errorx.Internal(c, "更新考勤记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "更新成功",
	})
}

// AdminDeleteAttendance godoc
// @Summary 管理员删除考勤记录
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Param id path int true "考勤记录ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/attendance/{id} [delete]
func AdminDeleteAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的考勤记录 ID", err)
		return
	}

	if err := dao.DeleteAttendance(uint(id64)); err != nil {
		errorx.Internal(c, "删除考勤记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除成功",
	})
}
