package handlers

import (
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ✅ 分页获取员工列表（仅管理员）
func GetAllPersons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.DefaultQuery("keyword", "")

	list, total, err := models.FetchPersonsPaged(page, pageSize, keyword)
	if err != nil {
		log.Printf("❌ 查询员工失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询员工失败",
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

// ✅ 根据 ID 查询员工详情（管理员可看任意，staff 仅可看自己）
func GetPersonByID(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("role")
	username, _ := c.Get("username")

	person, err := models.FetchPersonByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "未找到该员工",
		})
		return
	}

	// 🔐 staff 仅可查看本人；root/superadmin 直接放行
	if role == "staff" {
		acc, ok := models.GetAccountByUsername(fmt.Sprint(username))
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "无效的账号",
			})
			return
		}
		if acc.EmpID != person.EmpID {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 1,
				"msg":  "权限不足：无法查看他人信息",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": person,
	})
}

// ✅ 新增员工（仅管理员）
func CreatePerson(c *gin.Context) {
	var p models.Person

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求格式错误",
		})
		return
	}

	if p.Name == "" || p.DptID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "姓名或部门不能为空",
		})
		return
	}

	if err := models.CreatePerson(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "新增员工失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "新增成功",
	})
}

// ✅ 更新员工信息（管理员可更新任意，staff 仅可改本人，root 直接放行）
func UpdatePerson(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("role")
	username, _ := c.Get("username")

	var p models.Person
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求格式错误",
		})
		return
	}

	var acc models.Account
	var ok bool

	// 🧩 root/superadmin 账户直接跳过数据库验证
	if role == "superadmin" || fmt.Sprint(username) == "root" {
		acc = models.Account{Username: "root", Role: "superadmin", EmpID: "ROOT"}
		ok = true
	} else {
		acc, ok = models.GetAccountByUsername(fmt.Sprint(username))
	}

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "无效的账号",
		})
		return
	}

	targetPerson, err := models.FetchPersonByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 1,
			"msg":  "目标员工不存在",
		})
		return
	}

	// 🔐 staff 仅可修改本人信息；admin/superadmin 放行
	if role == "staff" && acc.EmpID != targetPerson.EmpID {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 1,
			"msg":  "权限不足：不能修改他人信息",
		})
		return
	}

	// ✅ 调用模型层更新（字段白名单在 models.UpdatePerson 控制）
	if err := models.UpdatePerson(id, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "更新失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "更新成功",
	})
}

// ✅ 删除员工（仅管理员）
func DeletePerson(c *gin.Context) {
	empID := c.Param("id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "EmpID 不能为空",
		})
		return
	}

	if err := models.DeletePerson(empID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "删除失败",
			"err":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除成功",
	})
}
