package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite" // SQLite 驱动
)

var DB *sql.DB // 全局数据库连接

// ==========================
// InitDB 初始化数据库
// ==========================
func InitDB() error {
	var err error

	// ================================
	// 确保数据库文件路径存在
	// ================================
	dbDir := "./db"
	if err = os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	dbPath := fmt.Sprintf("%s/hr.db", dbDir)

	// ================================
	// 高性能 & 稳定性参数
	// ================================
	connStr := fmt.Sprintf(
		"file:%s?_busy_timeout=5000&_foreign_keys=on&_journal_mode=WAL&_synchronous=NORMAL",
		dbPath,
	)

	DB, err = sql.Open("sqlite", connStr)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库不可用: %v", err)
	}
	log.Printf("✅ 数据库连接成功 (%s)\n", dbPath)

	// ================================
	// 全局 PRAGMA 设置（双保险）
	// ================================
	DB.Exec(`PRAGMA foreign_keys = ON;`)
	DB.Exec(`PRAGMA defer_foreign_keys = OFF;`)
	DB.Exec(`PRAGMA journal_mode = WAL;`)
	DB.Exec(`PRAGMA synchronous = NORMAL;`)

	// ================================
	// Schema 版本控制
	// ================================
	if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS META (version INTEGER DEFAULT 1)`); err != nil {
		return fmt.Errorf("创建 META 表失败: %v", err)
	}

	var version int
	_ = DB.QueryRow(`SELECT version FROM META LIMIT 1`).Scan(&version)
	if version == 0 {
		version = 1
		DB.Exec(`INSERT INTO META (version) VALUES (1)`)
	}
	// log.Printf("📦 当前数据库版本: %d\n", version)

	// ================================
	// PERSON 表
	// ================================
	createPerson := `
	CREATE TABLE IF NOT EXISTS PERSON (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		emp_id TEXT UNIQUE,
		auth INTEGER,
		name TEXT,
		sex TEXT,
		birth TEXT,
		dpt_id INTEGER,
		job TEXT,
		addr TEXT,
		tel TEXT,
		email TEXT,
		state INTEGER DEFAULT 1,
		remark TEXT,
		create_at TEXT DEFAULT (datetime('now')),
		update_at TEXT DEFAULT (datetime('now')),
		FOREIGN KEY (dpt_id) REFERENCES DEPARTMENT(id)
	);
	`
	if _, err := DB.Exec(createPerson); err != nil {
		return fmt.Errorf("创建 PERSON 表失败: %v", err)
	}

	// ================================
	// ACCOUNT 表
	// ================================
	createAccount := `
	CREATE TABLE IF NOT EXISTS ACCOUNT (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT,
		emp_id TEXT,
		role TEXT DEFAULT 'staff',
		status INTEGER DEFAULT 1,
		create_at TEXT DEFAULT (datetime('now')),
		last_login TEXT
	);
	`
	if _, err := DB.Exec(createAccount); err != nil {
		return fmt.Errorf("创建 ACCOUNT 表失败: %v", err)
	}

	// ================================
	// DEPARTMENT 表
	// ================================
	createDepartment := `
	CREATE TABLE IF NOT EXISTS DEPARTMENT (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE,
		manager TEXT,
		intro TEXT,
		create_at TEXT DEFAULT (datetime('now')),
		location TEXT,
		dpt_num INTEGER DEFAULT 0,
		full_num INTEGER DEFAULT 0,
		is_full INTEGER DEFAULT 0
	);
	`
	if _, err := DB.Exec(createDepartment); err != nil {
		return fmt.Errorf("创建 DEPARTMENT 表失败: %v", err)
	}

	// ✅ 确保存在默认部门
	var deptCount int
	_ = DB.QueryRow(`SELECT COUNT(*) FROM DEPARTMENT`).Scan(&deptCount)
	if deptCount == 0 {
		_, err := DB.Exec(`
			INSERT INTO DEPARTMENT (name, manager, intro, location, dpt_num, full_num, is_full)
			VALUES ('未分配部门', '系统', '系统默认部门', '未知', 0, 50, 0)
		`)
		if err != nil {
			return fmt.Errorf("创建默认部门失败: %v", err)
		}
		log.Println("📦 已自动创建默认部门: '未分配部门'")
	}

	// ================================
	// PERSONNEL 表（含 approver 字段）
	// ================================
	createPersonnel := `
	CREATE TABLE IF NOT EXISTS PERSONNEL (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		emp_id TEXT,
		change_type INTEGER,
		description TEXT,
		state INTEGER DEFAULT 0,
		target_dpt INTEGER,
		approver TEXT,
		create_at TEXT DEFAULT (datetime('now')),
		approve_at TEXT,
		FOREIGN KEY (emp_id) REFERENCES PERSON(emp_id)
	);
	`
	if _, err := DB.Exec(createPersonnel); err != nil {
		return fmt.Errorf("创建 PERSONNEL 表失败: %v", err)
	}

	// ================================
	// NOTICE 表
	// ================================
	createNotice := `
	CREATE TABLE IF NOT EXISTS NOTICE (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		content TEXT,
		publisher TEXT,
		create_at TEXT DEFAULT (datetime('now')),
		update_at TEXT
	);
	`
	if _, err := DB.Exec(createNotice); err != nil {
		return fmt.Errorf("创建 NOTICE 表失败: %v", err)
	}

	log.Println("✅ 所有表检查/创建完成")
	return nil
}

// ==========================
// CloseDB 优雅关闭数据库
// ==========================
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("⚠️ 关闭数据库出错: %v\n", err)
		} else {
			log.Println("✅ 数据库连接已关闭")
		}
	}
}
