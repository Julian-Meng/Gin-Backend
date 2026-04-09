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

// 档案返回结构体（组合 Account + Person + Department）

type DepartmentInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProfileResponse struct {
	ID        uint            `json:"id"`
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

// 工具：根据 Person 拼出 Profile（可选带账号信息）

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
		ID:        person.ID,
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

// 1. 当前登录用户查看自己的档案
// GET /api/user/profile
// 依赖：JWT 中间件已在 context 中设置 "username"

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

// 2. 管理员查看指定员工档案（含部门信息）
// GET /api/admin/person/profile/:emp_id

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

type UpdateMyProfileReq struct {
	Name   *string    `json:"name"`
	Sex    *string    `json:"sex"`
	Birth  *time.Time `json:"birth"`
	Job    *string    `json:"job"`
	Addr   *string    `json:"addr"`
	Tel    *string    `json:"tel"`
	Email  *string    `json:"email"`
	Remark *string    `json:"remark"`
}

// 3. 当前登录用户更新自己的档案（支持部分字段）
// PUT /api/user/profile
func UpdateMyProfile(c *gin.Context) {
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

	var req UpdateMyProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求解析失败",
			"err":  err.Error(),
		})
		return
	}

	account, ok := dao.GetAccountByUsername(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到当前用户对应的账号记录",
		})
		return
	}

	person, err := dao.FetchPersonByEmpID(account.EmpID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到当前用户对应的员工档案",
			"err":  err.Error(),
		})
		return
	}

	updates := map[string]interface{}{}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "name 不能为空"})
			return
		}
		updates["name"] = name
	}

	if req.Sex != nil {
		updates["sex"] = strings.TrimSpace(*req.Sex)
	}

	if req.Birth != nil {
		if req.Birth.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "出生日期不能晚于当前时间"})
			return
		}
		updates["birth"] = req.Birth
	}

	if req.Job != nil {
		updates["job"] = strings.TrimSpace(*req.Job)
	}

	if req.Addr != nil {
		updates["addr"] = strings.TrimSpace(*req.Addr)
	}

	if req.Tel != nil {
		updates["tel"] = strings.TrimSpace(*req.Tel)
	}

	if req.Email != nil {
		updates["email"] = strings.TrimSpace(*req.Email)
	}

	if req.Remark != nil {
		updates["remark"] = strings.TrimSpace(*req.Remark)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "至少提供一个可更新字段"})
		return
	}

	updates["updated_at"] = time.Now()

	if err := dao.UpdatePersonFields(person.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "更新档案失败",
			"err":  err.Error(),
		})
		return
	}

	updatedPerson, err := dao.FetchPersonByID(person.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "档案更新后读取失败",
			"err":  err.Error(),
		})
		return
	}

	profile, err := buildProfile(updatedPerson, &account)
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
		"msg":  "档案更新成功",
		"data": profile,
	})
}
