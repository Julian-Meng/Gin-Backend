package handlers

import (
	"backend/dao"
	"backend/db"
	"backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// =============================
// 档案返回结构体（组合 Account + Person + Department）
// =============================

type DepartmentInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProfileResponse struct {
	EmpID     string          `json:"emp_id"`
	Username  string          `json:"username,omitempty"`
	Role      string          `json:"role,omitempty"`
	Name      string          `json:"name"`
	Auth      int             `json:"auth"`
	Sex       string          `json:"sex"`
	Birth     *time.Time      `json:"birth,omitempty"`
	DptID     uint            `json:"dpt_id"`
	Job       string          `json:"job"`
	Addr      string          `json:"addr"`
	Tel       string          `json:"tel"`
	Email     string          `json:"email"`
	State     int             `json:"state"`
	Remark    string          `json:"remark"`
	DeptInfo  *DepartmentInfo `json:"department,omitempty"`
	CreatedAt time.Time       `json:"create_at"`
	UpdatedAt time.Time       `json:"update_at"`
}

// =============================
// 工具：根据 Person 拼出 Profile（可选带账号信息）
// =============================

func buildProfile(person *models.Person, account *models.Account) (*ProfileResponse, error) {
	var deptInfo *DepartmentInfo

	// 查部门信息（如果有部门）
	if person.DptID != 0 {
		dbConn := db.GetDB()
		var dept models.Department
		if err := dbConn.First(&dept, person.DptID).Error; err == nil {
			deptInfo = &DepartmentInfo{
				ID:   dept.ID,
				Name: dept.Name,
			}
		}
	}

	resp := &ProfileResponse{
		EmpID:     person.EmpID,
		Name:      person.Name,
		Auth:      person.Auth,
		Sex:       person.Sex,
		Birth:     person.Birth,
		DptID:     person.DptID,
		Job:       person.Job,
		Addr:      person.Addr,
		Tel:       person.Tel,
		Email:     person.Email,
		State:     person.State,
		Remark:    person.Remark,
		DeptInfo:  deptInfo,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}

	// 如果传入了账号信息，则一起带回去
	if account != nil {
		resp.Username = account.Username
		resp.Role = account.Role
	}

	return resp, nil
}

// =============================
// 1. 当前登录用户查看自己的档案
// GET /api/user/profile
// 依赖：JWT 中间件已在 context 中设置 "username"
// =============================

func GetMyProfile(c *gin.Context) {
	// 从 JWT 中拿 username
	val, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "未获取到登录用户信息",
		})
		return
	}
	username, _ := val.(string)
	username = strings.TrimSpace(username)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "登录用户名无效",
		})
		return
	}

	// 1) 查账号 Account（拿 EmpID）
	account, ok := dao.GetAccountByUsername(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到当前用户对应的账号记录",
		})
		return
	}

	// 2) 查员工 Person（根据 EmpID）
	person, err := dao.FetchPersonByEmpID(account.EmpID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到当前用户对应的员工档案",
			"err":  err.Error(),
		})
		return
	}

	// 3) 组合档案信息
	profile, err := buildProfile(person, &account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "构建档案信息失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": profile,
	})
}

// =============================
// 2. 管理员查看指定员工档案（含部门信息）
// GET /api/admin/person/profile/:emp_id
// =============================

func GetPersonProfile(c *gin.Context) {
	empID := strings.TrimSpace(c.Param("emp_id"))
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "emp_id 不能为空",
		})
		return
	}

	// 1) 查员工 Person
	person, err := dao.FetchPersonByEmpID(empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到对应的员工档案",
			"err":  err.Error(),
		})
		return
	}

	// 2) 尝试查账号（有些员工可能还没账号）
	var account *models.Account
	{
		dbConn := db.GetDB()
		var acc models.Account
		if err := dbConn.
			Where("emp_id = ?", empID).
			First(&acc).Error; err == nil {
			account = &acc
		}
	}

	// 3) 组合档案信息
	profile, err := buildProfile(person, account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "构建档案信息失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": profile,
	})
}
