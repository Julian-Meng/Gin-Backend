package models

import (
	"backend/db"
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// ==========================
// Account 表结构
// ==========================
type Account struct {
	ID        int            `json:"id"`
	Username  string         `json:"username"`
	Password  string         `json:"password,omitempty"`
	EmpID     string         `json:"emp_id"`
	Role      string         `json:"role"`
	Status    int            `json:"status"`
	CreateAt  string         `json:"create_at"`
	LastLogin sql.NullString `json:"last_login"`
}

// ==========================
// 注册新用户（自动生成 emp_id + 对应人员记录）
// ==========================
func InsertAccount(a Account) error {
	if a.Role == "" {
		a.Role = "staff"
	}
	if a.Password == "" {
		a.Password = "123456" // 默认密码
	}

	tx, txErr := db.DB.Begin()
	if txErr != nil {
		return fmt.Errorf("开启事务失败: %v", txErr)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if txErr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// ========================
	// 自动保证存在默认部门
	// ========================
	var defaultDptID int
	err := tx.QueryRow(`SELECT id FROM DEPARTMENT ORDER BY id LIMIT 1`).Scan(&defaultDptID)
	if err == sql.ErrNoRows || defaultDptID == 0 {
		res, createErr := tx.Exec(`
			INSERT INTO DEPARTMENT (name, manager, intro, location, dpt_num, full_num, is_full)
			VALUES ('未分配部门', '系统', '系统默认部门', '未知', 0, 50, 0)
		`)
		if createErr != nil {
			txErr = fmt.Errorf("创建默认部门失败: %v", createErr)
			return txErr
		}
		newID, _ := res.LastInsertId()
		defaultDptID = int(newID)
		log.Println("📦 自动创建默认部门成功: 未分配部门")
	} else if err != nil {
		txErr = fmt.Errorf("查询默认部门失败: %v", err)
		return txErr
	}

	// ========================
	// 生成员工号 EMP0001...
	// ========================
	var nextID int
	if err := tx.QueryRow(`SELECT IFNULL(MAX(id), 0) + 1 FROM ACCOUNT`).Scan(&nextID); err != nil {
		txErr = fmt.Errorf("生成 emp_id 失败: %v", err)
		return txErr
	}
	a.EmpID = fmt.Sprintf("EMP%04d", nextID)

	// ========================
	// 密码加密
	// ========================
	hashed, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		txErr = fmt.Errorf("密码加密失败: %v", err)
		return txErr
	}

	// ========================
	// 插入 ACCOUNT 表
	// ========================
	if _, err := tx.Exec(`
		INSERT INTO ACCOUNT (username, password, emp_id, role, status, create_at)
		VALUES (?, ?, ?, ?, 1, datetime('now'))
	`, a.Username, hashed, a.EmpID, a.Role); err != nil {
		txErr = fmt.Errorf("插入 ACCOUNT 失败: %v", err)
		return txErr
	}

	// ========================
	// 自动插入 PERSON 记录
	// ========================
	var personCount int
	_ = tx.QueryRow(`SELECT COUNT(*) FROM PERSON`).Scan(&personCount)
	defaultName := fmt.Sprintf("用户%d", personCount+1)

	if _, err := tx.Exec(`
		INSERT INTO PERSON (emp_id, name, auth, sex, birth, dpt_id, job, addr, tel, email, state, remark, create_at, update_at)
		VALUES (?, ?, 0, '', '', ?, '', '', '', '', 1, '', datetime('now'), datetime('now'))
	`, a.EmpID, defaultName, defaultDptID); err != nil {
		txErr = fmt.Errorf("插入 PERSON 失败: %v", err)
		return txErr
	}

	return nil
}

// ==========================
// 登录验证
// ==========================
func ValidateLogin(username, password string) (Account, bool) {
	var acc Account
	err := db.DB.QueryRow(`
		SELECT id, username, password, emp_id, role, status, create_at, last_login
		FROM ACCOUNT WHERE username = ?
	`, username).Scan(&acc.ID, &acc.Username, &acc.Password, &acc.EmpID, &acc.Role, &acc.Status, &acc.CreateAt, &acc.LastLogin)

	if err == sql.ErrNoRows {
		return Account{}, false
	}
	if err != nil {
		log.Println("数据库查询错误:", err)
		return Account{}, false
	}

	// 密码比对
	if bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)) != nil {
		return Account{}, false
	}

	// 更新登录时间
	_, _ = db.DB.Exec(`UPDATE ACCOUNT SET last_login = datetime('now') WHERE id = ?`, acc.ID)
	acc.Password = ""
	return acc, true
}

// ==========================
// 获取所有账号列表
// ==========================
func GetAllAccounts() []Account {
	rows, err := db.DB.Query(`
		SELECT id, username, emp_id, role, status, create_at, last_login
		FROM ACCOUNT
		ORDER BY id DESC
	`)
	if err != nil {
		log.Println("查询账号列表错误:", err)
		return nil
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Username, &a.EmpID, &a.Role, &a.Status, &a.CreateAt, &a.LastLogin); err == nil {
			accounts = append(accounts, a)
		}
	}
	return accounts
}

// ==========================
// 更新账号信息
// ==========================
func UpdateAccount(id string, a Account) error {
	// 只允许 admin 或 staff
	if a.Role != "admin" && a.Role != "staff" {
		return fmt.Errorf("无效的角色类型: %s，只允许 admin 或 staff", a.Role)
	}

	_, err := db.DB.Exec(`
		UPDATE ACCOUNT
		SET role = ?, status = ?, last_login = datetime('now')
		WHERE id = ?
	`, a.Role, a.Status, id)
	if err != nil {
		log.Printf("❌ 更新账号失败: %v", err)
	}
	return err
}

// ==========================
// 根据用户名查询账号信息
// ==========================
func GetAccountByUsername(username string) (Account, bool) {
	var acc Account
	err := db.DB.QueryRow(`
		SELECT id, username, emp_id, role, status 
		FROM ACCOUNT WHERE username = ?
	`, username).Scan(&acc.ID, &acc.Username, &acc.EmpID, &acc.Role, &acc.Status)
	if err != nil {
		return Account{}, false
	}
	return acc, true
}

// ==========================
// 删除账号
// ==========================
func DeleteAccount(id string) error {
	_, err := db.DB.Exec(`DELETE FROM ACCOUNT WHERE id = ?`, id)
	return err
}
