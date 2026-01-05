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

// 数据库配置结构体
type Config struct {
	Driver string
	DSN    string
	Debug  bool
	DBName string // MySQL 自动建库
}

// MySQL 自动创建数据库
func ensureMySQLDatabase(cfg Config) error {
	if cfg.Driver != "mysql" || cfg.DBName == "" {
		return nil
	}

	// 解析 DSN（去掉数据库名）
	dsnWithoutDB := cfg.DSN
	if idx := len(cfg.DBName); idx > 0 {
		dsnWithoutDB = cfg.DSN[:len(cfg.DSN)-idx]
	}

	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;", cfg.DBName)

	tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("MySQL 建库连接失败: %v", err)
	}

	if err := tempDB.Exec(sql).Error; err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	fmt.Println("MySQL 数据库已确保存在 →", cfg.DBName)
	return nil
}

// InitDB - 初始化数据库
func InitDB(cfg Config) error {

	// 先尝试自动建库（MySQL）
	if cfg.Driver == "mysql" {
		if err := ensureMySQLDatabase(cfg); err != nil {
			return err
		}
	}

	// 设置 GORM 日志级别
	var gLogger logger.Interface
	if cfg.Debug {
		gLogger = logger.Default.LogMode(logger.Info)
	} else {
		gLogger = logger.Default.LogMode(logger.Warn)
	}

	var db *gorm.DB
	var err error

	switch cfg.Driver {
	case "sqlite":
		ensureSQLiteDir()
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{Logger: gLogger})

	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{Logger: gLogger})

	default:
		return fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	DB = db
	log.Printf("\033[32m数据库连接成功 → [%s]\033[0m", cfg.Driver)

	// MySQL 设置连接池
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// 自动迁移
	if err := autoMigrateAll(); err != nil {
		return fmt.Errorf("auto migrate failed: %v", err)
	}

	// 默认部门
	if err := ensureDefaultDepartment(); err != nil {
		return fmt.Errorf("ensure default department failed: %v", err)
	}

	return nil
}

// AutoMigrate 所有模型
func autoMigrateAll() error {
	return DB.AutoMigrate(
		&models.Account{},
		&models.Person{},
		&models.Department{},
		&models.Personnel{},
		&models.Notice{},
		&models.Attendance{},
	)
}

// SQLite 目录创建（移除未使用参数警告）
func ensureSQLiteDir() {
	dir := "db"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// 创建默认部门
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
		log.Println("\033[32m已自动创建默认部门: 未分配部门\033[0m")
	}
	return nil
}

// 返回全局 DB
func GetDB() *gorm.DB { return DB }

// 关闭数据库连接
func CloseDB() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}
