package main

import (
	"backend/db"
	"backend/models"
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
	//明文秘钥，生产环境请使用环境变量
	os.Setenv("JWT_SECRET", "MyProductionSecretKey")

	start := time.Now()
	// 设置gin模式为release模式,控制台不打印日志
	gin.SetMode(gin.ReleaseMode)

	// 初始化数据库
	if err := db.InitDB(); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	if err := models.EnsureDefaultDepartment(); err != nil {
		log.Fatalf("❌ 检查默认部门失败: %v", err)
	}

	// 创建 Gin 路由
	r := SetupRouter()

	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Println("✅ 服务正在启动，请稍候...")

	// 异步启动 HTTP 服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()

	elapsed := time.Since(start)
	log.Printf("✅ 服务已启动在 http://localhost%s", addr)
	log.Printf("⏱️ 启动耗时: %v", elapsed)

	// 捕获系统信号（Ctrl + C / kill / systemctl stop）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // ⏸ 阻塞等待信号

	log.Println("✅ 收到退出信号，正在安全关闭服务...")

	// 设置超时上下文，防止卡死
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭 HTTP 服务
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ 服务关闭过程中出错: %v", err)
	} else {
		log.Println("✅ HTTP 服务已安全关闭")
	}

	// 关闭数据库
	db.CloseDB()
}
