package handlers

import (
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ======================================
// 用户或管理员创建人事变更申请
// ======================================
func CreatePersonnelChange(c *gin.Context) {
	var req models.Personnel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求格式错误",
		})
		return
	}

	if req.EmpID == "" || req.TargetDpt == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "员工ID或目标部门不能为空",
		})
		return
	}

	if err := models.CreatePersonnelChange(req); err != nil {
		log.Println("❌ 变更申请失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "变更申请失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "变更申请已提交",
	})
}

// ======================================
// 分页查询所有人事变更记录（管理员）
// ======================================
func GetAllPersonnelChanges(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	list, total, err := models.FetchPersonnelPaged(limit, offset)
	if err != nil {
		log.Println("❌ 查询变更记录失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "查询失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"list":  list,
			"total": total,
		},
	})
}

// ======================================
// 获取单条人事变更详情
// ======================================
func GetPersonnelChangeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "ID格式错误"})
		return
	}

	p, err := models.GetPersonnelByID(id)
	if err != nil {
		log.Println("❌ 获取变更详情失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "查询失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": p,
	})
}

// ======================================
// 管理员审批人事变更
// ======================================
func ApprovePersonnelChange(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "ID格式错误"})
		return
	}

	var req struct {
		Approve  bool   `json:"approve"`
		Approver string `json:"approver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求格式错误",
		})
		return
	}

	if req.Approver == "" {
		req.Approver = "系统管理员"
	}

	if err := models.ApprovePersonnelChange(id, req.Approver, req.Approve); err != nil {
		log.Println("❌ 审批失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "审批失败",
			"error": err.Error(),
		})
		return
	}

	status := "已驳回"
	if req.Approve {
		status = "已通过"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "审批成功",
		"data": gin.H{
			"id":       id,
			"status":   status,
			"approver": req.Approver,
		},
	})
}

// ======================================
// 用户查看自己的人事变更记录（可选）
// ======================================
func GetUserPersonnelChanges(c *gin.Context) {
	empID := c.Query("emp_id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "缺少 emp_id 参数"})
		return
	}

	rows, _, err := models.FetchPersonnelPaged(100, 0) // ✅ 修正这里
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "查询失败",
			"error": err.Error(),
		})
		return
	}

	// 过滤出该用户的记录
	var userChanges []models.Personnel
	for _, p := range rows {
		if p.EmpID == empID {
			userChanges = append(userChanges, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"list":  userChanges,
			"total": len(userChanges),
		},
	})
}
