package main

import (
	"backend/db"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	// ===============================
	// Gin 运行模式（减少日志输出）
	// 可选：gin.DebugMode / gin.ReleaseMode / gin.TestMode
	// ===============================
	gin.SetMode(gin.ReleaseMode) // 正式环境推荐使用，日志更干净
	// gin.SetMode(gin.DebugMode) // 开发时可手动开启

	// ===============================
	// JWT Secret
	// ===============================
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "MyPrivateSecretKey")
	}

	// ===============================
	// 启动时间统计（开始计时）
	// ===============================
	start := time.Now()

	// ===============================
	// 数据库配置：SQLite（默认）
	// ===============================
	cfg := db.Config{
		Driver: "sqlite",     // "mysql" 或 "sqlite"
		DSN:    "./db/hr.db", // SQLite 路径
		Debug:  false,        // 关闭 SQL 日志（若你要更干净的输出）
	}

	// ===============================
	//  MySQL 示例配置
	// ===============================

	// cfg := db.Config{
	// 	Driver: "mysql",
	// 	DSN:    "user:123@tcp(localhost:3306)/hrdb?charset=utf8mb4&parseTime=True&loc=Local",
	// 	Debug:  false,
	// }

	// ===============================
	// 初始化数据库
	// ===============================
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// ===============================
	// 初始化路由
	// ===============================
	r := SetupRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// ===============================
	// 启动服务
	// ===============================
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()

	// ===============================
	// 启动完成 → 打印耗时
	// ===============================
	bootCost := time.Since(start)
	log.Println("✅ 服务已启动 → http://localhost:8080")
	log.Printf("✅ 启动耗时: %.2f 秒\n", bootCost.Seconds())

	// ===============================
	// 退出
	// ===============================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("✅ 正在停止服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("❌ 关闭服务时出错:", err)
	}

	// ===============================
	// 关闭数据库连接
	// ===============================
	db.CloseDB()

	log.Println("✅ 服务已安全退出")
}
