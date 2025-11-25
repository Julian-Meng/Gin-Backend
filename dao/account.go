package dao

import (
	"backend/db"
	"backend/models"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ==========================
// 内部工具：生成 EMP 编号
// ==========================
func generateEmpID(tx *gorm.DB) (string, error) {
	var maxID int64
	if err := tx.Table("ACCOUNT").Select("IFNULL(MAX(id),0)").Scan(&maxID).Error; err != nil {
		return "", fmt.Errorf("生成 emp_id 失败: %v", err)
	}
	return fmt.Sprintf("EMP%04d", maxID+1), nil
}

// ==========================
// 注册新用户（自动生成 emp_id + Person）
// ==========================
func InsertAccount(a models.Account) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {
		// 1. 确保存在一个默认部门（未分配部门）
		var defaultDept models.Department
		err := tx.Order("id").First(&defaultDept).Error
		if err == gorm.ErrRecordNotFound {
			defaultDept = models.Department{
				Name:     "未分配部门",
				Manager:  "系统",
				Intro:    "系统默认部门",
				Location: "未知",
				DptNum:   0,
				FullNum:  50,
				IsFull:   0,
			}
			if err := tx.Create(&defaultDept).Error; err != nil {
				return fmt.Errorf("创建默认部门失败: %v", err)
			}
			log.Println("📦 自动创建默认部门成功: 未分配部门")
		} else if err != nil {
			return fmt.Errorf("查询默认部门失败: %v", err)
		}

		// 2. 生成员工号 EMP0001...
		empID, err := generateEmpID(tx)
		if err != nil {
			return err
		}
		a.EmpID = empID

		// 3. 默认角色 & 密码
		if a.Role == "" {
			a.Role = "staff"
		}
		if a.Password == "" {
			a.Password = "123456"
		}

		// 4. 密码加密
		hashed, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败: %v", err)
		}
		a.Password = string(hashed)
		a.Status = 1

		// 5. 插入 ACCOUNT
		if err := tx.Create(&a).Error; err != nil {
			return fmt.Errorf("插入 ACCOUNT 失败: %v", err)
		}

		// 6. 自动插入 PERSON 记录
		var personCount int64
		tx.Model(&models.Person{}).Count(&personCount)
		defaultName := fmt.Sprintf("用户%d", personCount+1)

		p := models.Person{
			EmpID:  a.EmpID,
			Name:   defaultName,
			Auth:   0,
			Sex:    "",
			Birth:  nil, // 如果你之后把 Birth 改成 time.Time，这里再一起升级
			DptID:  uint(defaultDept.ID),
			Job:    "",
			Addr:   "",
			Tel:    "",
			Email:  "",
			State:  1,
			Remark: "",
			// CreatedAt / UpdatedAt 交给 GORM 也可以，这里就不手动填了
		}

		if err := tx.Create(&p).Error; err != nil {
			return fmt.Errorf("插入 PERSON 失败: %v", err)
		}

		return nil
	})
}

// ==========================
// 登录验证
// ==========================
func ValidateLogin(username, password string) (models.Account, bool) {
	dbConn := db.GetDB()

	var acc models.Account
	err := dbConn.Where("username = ?", username).First(&acc).Error
	if err == gorm.ErrRecordNotFound {
		return models.Account{}, false
	}
	if err != nil {
		log.Println("数据库查询错误:", err)
		return models.Account{}, false
	}

	// 密码比对
	if bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)) != nil {
		return models.Account{}, false
	}

	// 更新登录时间
	now := time.Now()
	dbConn.Model(&acc).Update("last_login", &now)

	acc.Password = "" // 不把密码返回给上层
	return acc, true
}

// ==========================
// 获取所有账号列表
// ==========================
func GetAllAccounts() []models.Account {
	dbConn := db.GetDB()

	var accounts []models.Account
	if err := dbConn.
		Order("id DESC").
		Find(&accounts).Error; err != nil {
		log.Println("查询账号列表错误:", err)
		return nil
	}
	return accounts
}

// ==========================
// 更新账号信息（角色 / 状态）
// ==========================
func UpdateAccount(id string, a models.Account) error {
	// 只允许 admin 或 staff
	if a.Role != "admin" && a.Role != "staff" {
		return fmt.Errorf("无效的角色类型: %s，只允许 admin 或 staff", a.Role)
	}

	dbConn := db.GetDB()

	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的账号 ID: %s", id)
	}

	updates := map[string]interface{}{
		"role":       a.Role,
		"status":     a.Status,
		"last_login": time.Now(),
	}

	if err := dbConn.Model(&models.Account{}).
		Where("id = ?", uid).
		Updates(updates).Error; err != nil {
		log.Printf("❌ 更新账号失败: %v", err)
		return err
	}
	return nil
}

// ==========================
// 根据用户名查询账号信息
// ==========================
func GetAccountByUsername(username string) (models.Account, bool) {
	dbConn := db.GetDB()

	var acc models.Account
	err := dbConn.
		Select("id, username, emp_id, role, status").
		Where("username = ?", username).
		First(&acc).Error
	if err != nil {
		return models.Account{}, false
	}
	return acc, true
}

// ==========================
// 删除账号
// ==========================
func DeleteAccount(id string) error {
	dbConn := db.GetDB()

	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的账号 ID: %s", id)
	}

	return dbConn.Delete(&models.Account{}, uid).Error
}
