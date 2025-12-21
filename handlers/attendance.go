package handlers

import (
	"backend/dao"
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

// 1. 用户签到
// POST /api/user/attendance/checkin

func UserCheckIn(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号未绑定员工档案，无法签到",
		})
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号 emp_id 无效，无法签到",
		})
		return
	}

	if err := dao.CheckIn(empID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "签到成功",
	})
}

// 2. 用户签退
// POST /api/user/attendance/checkout

func UserCheckOut(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号未绑定员工档案，无法签退",
		})
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号 emp_id 无效，无法签退",
		})
		return
	}

	if err := dao.CheckOut(empID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "签退成功",
	})
}

// 3. 用户查看自己的考勤记录
// GET /api/user/attendance/my?start=2025-01-01&end=2025-01-31&page=1&pageSize=10

func GetMyAttendance(c *gin.Context) {
	val, exists := c.Get("emp_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号未绑定员工档案，无法查询考勤",
		})
		return
	}
	empID, _ := val.(string)
	empID = strings.TrimSpace(empID)
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "当前账号 emp_id 无效，无法查询考勤",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询考勤记录失败",
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

// 4. 管理员：按条件搜索考勤记录
// GET /api/admin/attendance?emp_id=EMP0001&dpt_id=3&start=2025-01-01&end=2025-01-31&page=1&pageSize=10

func AdminSearchAttendance(c *gin.Context) {
	empID := strings.TrimSpace(c.Query("emp_id"))
	dptIDStr := strings.TrimSpace(c.Query("dpt_id"))

	var dptID uint64
	var err error
	if dptIDStr != "" {
		dptID, err = strconv.ParseUint(dptIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 1,
				"msg":  "无效的 dpt_id",
			})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询考勤记录失败",
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

// 5. 管理员：更新考勤记录
// PUT /api/admin/attendance/:id
// Body 示例：{
//   "status": 2,
//   "remark": "迟到 10 分钟",
//   "check_in": "2025-01-01T09:10:00+08:00",
//   "check_out": "2025-01-01T18:02:00+08:00"
// }

func AdminUpdateAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的考勤记录 ID",
		})
		return
	}

	var req struct {
		Status   *int       `json:"status"`
		Remark   *string    `json:"remark"`
		CheckIn  *time.Time `json:"check_in"`
		CheckOut *time.Time `json:"check_out"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求体解析失败",
			"err":  err.Error(),
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "没有任何可更新的字段",
		})
		return
	}

	if err := dao.UpdateAttendance(uint(id64), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "更新考勤记录失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "更新成功",
	})
}

// 6. 管理员：删除考勤记录
// DELETE /api/admin/attendance/:id

func AdminDeleteAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的考勤记录 ID",
		})
		return
	}

	if err := dao.DeleteAttendance(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除考勤记录失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除成功",
	})
}
