package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/models"
)

var DB *gorm.DB

// ===============================
// 数据库配置结构体（env → main.go 传入）
// ===============================
type Config struct {
	Driver string // mysql / sqlite
	DSN    string // MySQL: user:pass@tcp(...) ; SQLite: ./db/hr.db
	Debug  bool   // 是否开启 GORM debug 模式
}

// ===============================
// InitDB - 初始化数据库（MySQL / SQLite）
// ===============================
func InitDB(cfg Config) error {

	// 设置 GORM 日志级别
	var gLogger logger.Interface
	if cfg.Debug {
		gLogger = logger.Default.LogMode(logger.Info)
	} else {
		gLogger = logger.Default.LogMode(logger.Warn)
	}

	// 初始化连接
	var (
		db  *gorm.DB
		err error
	)

	switch cfg.Driver {
	case "sqlite":
		ensureSQLiteDir(cfg.DSN)
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
			Logger: gLogger,
		})

	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
			Logger: gLogger,
		})

	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	DB = db
	log.Printf("✅ 数据库连接成功 → [%s]", cfg.Driver)

	// 设置连接池（MySQL 专用，SQLite 忽略）
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ===========================================
	// AutoMigrate（可关闭，如不想 GORM 改表结构）
	// ===========================================
	err = autoMigrateAll()
	if err != nil {
		return fmt.Errorf("auto migrate failed: %v", err)
	}

	// ===========================================
	// 确保默认部门存在（从原逻辑迁移）
	// ===========================================
	err = ensureDefaultDepartment()
	if err != nil {
		return fmt.Errorf("failed ensuring default department: %v", err)
	}

	return nil
}

// ===============================
// AutoMigrate 所有模型
// ===============================
func autoMigrateAll() error {
	return DB.AutoMigrate(
		&models.Account{},
		&models.Person{},
		&models.Department{},
		&models.Personnel{},
		&models.Notice{},
		// Dashboard 无表，不参与
	)
}

// ===============================
// SQLite 目录创建
// ===============================
func ensureSQLiteDir(path string) {
	dir := "db"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// ===============================
// 创建默认部门（跨库兼容）
// ===============================
func ensureDefaultDepartment() error {
	var count int64
	DB.Model(&models.Department{}).Count(&count)
	if count == 0 {
		def := models.Department{
			Name:     "未分配部门",
			Manager:  "系统",
			Intro:    "系统默认部门",
			Location: "未知",
			DptNum:   0,
			FullNum:  50,
			IsFull:   0,
		}
		if err := DB.Create(&def).Error; err != nil {
			return fmt.Errorf("创建默认部门失败: %v", err)
		}
		log.Println("📦 已自动创建默认部门: 未分配部门")
	}
	return nil
}

// ===============================
// 返回全局 DB
// ===============================
func GetDB() *gorm.DB {
	return DB
}

// ===============================
// 关闭数据库连接
// ===============================
func CloseDB() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}
