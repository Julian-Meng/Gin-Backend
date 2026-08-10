package main

import (
	_ "backend/docs"
	"backend/internal/db"
	"backend/internal/handler"
	"context"
	"errors"
	"fmt"
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

// @title Gin Backend API
// @version 1.0
// @description 人事管理后端 API 文档（含 admin/user/chat 等接口）
// @description 联调页面：[Backend Test Console](/bt)
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer <token>
func main() {
	// 加载 .env
	if err := godotenv.Load(); err != nil {
		printEnvFileMissingAndExit()
	}
	// 初始化 Auth 配置读取.env
	handler.MustInitAuthConfig()

	// Gin Mode
	ginMode := strings.TrimSpace(getEnv("GIN_MODE", gin.ReleaseMode))
	gin.SetMode(ginMode)

	// JWT Secret
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("\033[31mJWT_SECRET 未设置：请在 .env 或系统环境变量中配置 JWT_SECRET\033[0m")
	}

	// 启动时间统计
	start := time.Now()

	// 数据库配置
	dbDriver := strings.TrimSpace(getEnv("DB_DRIVER", "sqlite"))
	dbDSN := strings.TrimSpace(getEnv("DB_DSN", "./internal/db/hr.db"))
	dbDebug := getEnvBool("DB_DEBUG", false)

	cfg := db.Config{
		Driver: dbDriver,
		DSN:    dbDSN,
		Debug:  dbDebug,
	}

	// 初始化数据库
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("\033[31m数据库初始化失败: %v\033[0m", err)
	}

	// 初始化路由
	r := SetupRouter()

	// Server 配置（地址/端口从环境变量读取）
	addr := strings.TrimSpace(getEnv("SERVER_ADDR", ":2077"))
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("\033[31m服务启动失败: %v\033[0m", err)
		}
	}()

	// 启动完成 → 打印耗时
	bootCost := time.Since(start)
	log.Printf("\033[32m服务运行于 → http://localhost%s\033[0m\n", server.Addr)
	log.Printf("\033[32m启动耗时: %d ms\033[0m\n", bootCost.Milliseconds())
	log.Printf("\033[32m后端模式: %s | 数据库: %s | 数据库Debug: %v\033[0m\n", ginMode, dbDriver, dbDebug)

	// 退出
	shutdownTimeoutSec := getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 5)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\033[33m正在停止服务...\033[0m")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeoutSec)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("\033[31m关闭服务时出错:\033[0m", err)
	}

	// 关闭数据库连接
	db.CloseDB()

	log.Println("\033[32m服务已安全退出\033[0m")
}

// printEnvFileMissingAndExit 在 .env 文件缺失时打印双语提示并等待用户确认后退出。
func printEnvFileMissingAndExit() {
	separator := strings.Repeat("=", 56)

	fmt.Println()
	fmt.Println(separator)
	log.Println("\033[31m  错误!  未找到 .env 配置文件\033[0m")
	log.Println("\033[31m  ERROR!  .env configuration file not found\033[0m")
	fmt.Println(separator)
	fmt.Println()
	fmt.Println("  请在应用程序同级目录下创建 .env 文件")
	fmt.Println("  并配置所需的环境变量。")
	fmt.Println()
	fmt.Println("  Please create a .env file in the same")
	fmt.Println("  directory as this application and configure")
	fmt.Println("  the required environment variables.")
	fmt.Println()
	fmt.Println(separator)
	fmt.Println("  参考配置项 / Required items:")
	fmt.Println("    JWT_SECRET=your_secret_key")
	fmt.Println("    SUPERADMIN_ENABLED=true")
	fmt.Println("    SUPERADMIN_USERNAME=admin")
	fmt.Println("    SUPERADMIN_PASSWORD=your_password")
	fmt.Println("    SUPERADMIN_ROLE=superadmin")
	fmt.Println("    DB_DRIVER=sqlite")
	fmt.Println("    DB_DSN=./internal/db/hr.db")
	fmt.Println(separator)
	fmt.Println()
	fmt.Print("  按 Enter 键退出 / Press Enter to exit...")
	fmt.Scanln()
	os.Exit(1)
}

// getEnv 获取环境变量，若未设置则返回默认值
func getEnv(key, defaultVal string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	return v
}

// getEnvBool 获取布尔类型的环境变量，若未设置或解析失败则返回默认值
func getEnvBool(key string, defaultVal bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf("\033[43m环境变量 %s=%q 不是合法 bool，将使用默认值 %v\033[0m\n", key, v, defaultVal)
		return defaultVal
	}
	return b
}

// getEnvInt 获取整数类型的环境变量，若未设置或解析失败则返回默认值
func getEnvInt(key string, defaultVal int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("\033[43m环境变量 %s=%q 不是合法 int，将使用默认值 %d\033[0m\n", key, v, defaultVal)
		return defaultVal
	}
	return i
}
