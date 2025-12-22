package main

import (
	"backend/db"
	"backend/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1) 加载 .env
	if err := godotenv.Load(); err != nil {
		// 不强制报错：允许你在生产环境不使用 .env
		log.Println("⚠️ 未找到 .env 文件")
	}
	// 初始化 Auth 配置读取.env
	handlers.MustInitAuthConfig()

	// 2) Gin Mode
	ginMode := strings.TrimSpace(getEnv("GIN_MODE", gin.ReleaseMode))
	gin.SetMode(ginMode)

	// 3) JWT Secret
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET 未设置：请在 .env 或系统环境变量中配置 JWT_SECRET")
	}

	// 4) 启动时间统计
	start := time.Now()

	// 5) 数据库配置
	dbDriver := strings.TrimSpace(getEnv("DB_DRIVER", "sqlite"))
	dbDSN := strings.TrimSpace(getEnv("DB_DSN", "./db/hr.db"))
	dbDebug := getEnvBool("DB_DEBUG", false)

	cfg := db.Config{
		Driver: dbDriver,
		DSN:    dbDSN,
		Debug:  dbDebug,
	}

	// 初始化数据库
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 6) 初始化路由
	r := SetupRouter()

	// 7) Server 配置（地址/端口从环境变量读取）
	addr := strings.TrimSpace(getEnv("SERVER_ADDR", ":2077"))
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()

	// 启动完成 → 打印耗时
	bootCost := time.Since(start)
	log.Printf("✅ 服务运行于 → http://localhost%s\n", server.Addr)
	log.Printf("✅ 启动耗时: %d ms\n", bootCost.Milliseconds())
	log.Printf("✅ 后端模式: %s | 数据库: %s | 数据库Debug: %v\n", ginMode, dbDriver, dbDebug)

	// 8) 优雅退出
	shutdownTimeoutSec := getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 5)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("✅ 正在停止服务...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeoutSec)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("❌ 关闭服务时出错:", err)
	}

	// 关闭数据库连接
	db.CloseDB()

	log.Println("✅ 服务已安全退出")
}

func getEnv(key, defaultVal string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	return v
}

func getEnvBool(key string, defaultVal bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf("⚠️ 环境变量 %s=%q 不是合法 bool，将使用默认值 %v\n", key, v, defaultVal)
		return defaultVal
	}
	return b
}

func getEnvInt(key string, defaultVal int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("⚠️ 环境变量 %s=%q 不是合法 int，将使用默认值 %d\n", key, v, defaultVal)
		return defaultVal
	}
	return i
}
