// router.go
package main

import (
	"backend/handlers"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化路由配置
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// ========================
	//  CORS 中间件（允许跨域）
	// ========================
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 开发阶段允许全部源,生产阶段需改为具体域名
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ========================
	//  静态文件
	// ========================
	// 指定静态资源目录（例如 /static 下放图片、CSS、JS）
	r.Static("/static", "./static")

	// 主页（测试控制台 HTML）
	r.StaticFile("/", "./static/full_backend_test_pro.html")

	// ========================
	//  公共接口（无需登录）
	// ========================
	r.POST("/api/login", handlers.Login)
	r.POST("/api/register", handlers.Register)
	r.GET("/api/notice", handlers.GetAllNotices)

	// ========================
	//  管理员接口（需管理员权限）
	// ========================
	admin := r.Group("/api/admin")
	admin.Use(middlewares.JWTAuthMiddleware(), middlewares.AdminOnly())
	{
		// 仪表盘
		admin.GET("/dashboard", handlers.AdminDashboard)

		// 员工管理
		admin.GET("/employees", handlers.GetAllPersons)
		admin.POST("/employee", handlers.CreatePerson)
		admin.PUT("/employee/:id", handlers.UpdatePerson)
		admin.DELETE("/employee/:id", handlers.DeletePerson)
		admin.GET("/employee/:id", handlers.GetPersonByID)

		// 部门管理
		admin.GET("/departments", handlers.GetAllDepartments)
		admin.GET("/department/:id", handlers.GetDepartmentByID)
		admin.POST("/department", handlers.CreateDepartment)
		admin.PUT("/department/:id", handlers.UpdateDepartment)
		admin.DELETE("/department/:id", handlers.DeleteDepartment)

		// 人事变更管理
		admin.GET("/changes", handlers.GetAllPersonnelChanges)
		admin.POST("/change", handlers.CreatePersonnelChange)
		admin.PUT("/changes/:id", handlers.ApprovePersonnelChange)

		// 账号管理
		admin.GET("/auth", handlers.GetAllAccounts)
		admin.PUT("/auth/:id", handlers.UpdateAccount)
		admin.DELETE("/auth/:id", handlers.DeleteAccount)

		// 公告管理
		admin.POST("/notice", handlers.CreateNotice)
		admin.PUT("/notice/:id", handlers.UpdateNotice)
		admin.DELETE("/notice/:id", handlers.DeleteNotice)
		admin.GET("/notice/:id", handlers.GetNoticeByID)
	}

	// ========================
	//  普通用户接口（需登录）
	// ========================
	user := r.Group("/api/user")
	user.Use(middlewares.JWTAuthMiddleware())
	{
		user.GET("/dashboard", handlers.UserDashboard)
		user.GET("/profile/:id", handlers.GetPersonByID)
		user.PUT("/profile/:id", handlers.UpdatePerson)
		user.GET("/department/:id", handlers.GetDepartmentByID)
		user.PUT("/account/:id", handlers.UpdateAccount)
		user.POST("/change/request", handlers.CreatePersonnelChange)
	}

	return r
}
