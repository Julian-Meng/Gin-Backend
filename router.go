package main

import (
	"backend/handlers"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/docs.html")
	r.StaticFile("/bt", "./static/backend_test.html")
	r.StaticFile("/jv", "./static/json_viewer.html")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 无需登录
	r.POST("/api/login", handlers.Login)
	r.POST("/api/register", handlers.Register)
	r.GET("/api/notice", handlers.GetAllNotices)
	r.POST("/api/chat", handlers.ChatWithAI)

	// 管理员接口
	admin := r.Group("/api/admin")
	admin.Use(middlewares.JWTAuthMiddleware(), middlewares.AdminOnly())
	{
		// 仪表盘
		admin.GET("/dashboard", handlers.AdminDashboard)

		// 员工管理（person.go）
		admin.GET("/persons", handlers.GetPersons)
		admin.POST("/person", handlers.CreatePerson)
		admin.PUT("/person/:id", handlers.UpdatePerson)
		admin.DELETE("/person/:id", handlers.DeletePersonByID)
		admin.DELETE("/person/emp/:emp_id", handlers.DeletePersonByEmpID)
		admin.GET("/person/:id", handlers.GetPersonByID)
		admin.PUT("/person/job", handlers.ChangePersonJob)
		admin.PUT("/person/state", handlers.ChangePersonState)
		admin.PUT("/person/change-dept", handlers.ChangePersonDepartment)

		// 部门管理（department.go）
		admin.GET("/departments", handlers.GetDepartments)
		admin.GET("/department/:id", handlers.GetDepartmentByID)
		admin.POST("/department", handlers.CreateDepartment)
		admin.PUT("/department/:id", handlers.UpdateDepartment)
		admin.DELETE("/department/:id", handlers.DeleteDepartment)

		// 人事变更（personnel.go）
		admin.GET("/changes", handlers.GetPersonnelList)
		admin.GET("/change/:id", handlers.GetPersonnelByID)
		admin.POST("/change", handlers.CreatePersonnel)
		admin.PUT("/change/approve", handlers.ApprovePersonnel)

		// 账号管理（account.go）
		admin.GET("/accounts", handlers.GetAllAccounts)
		admin.POST("/account", handlers.CreateAccount)
		admin.PUT("/account/:id", handlers.UpdateAccount)
		admin.DELETE("/account/:id", handlers.DeleteAccount)

		// 公告管理（notice.go）
		admin.POST("/notice", handlers.CreateNotice)
		admin.PUT("/notice/:id", handlers.UpdateNotice)
		admin.DELETE("/notice/:id", handlers.DeleteNotice)
		admin.GET("/notice/:id", handlers.GetNoticeByID)

		// 个人档案（profile.go）
		admin.GET("/person/profile/:emp_id", handlers.GetPersonProfile)

		// 考勤管理（attendance.go）
		admin.GET("/attendance", handlers.AdminSearchAttendance)
		admin.PUT("/attendance/:id", handlers.AdminUpdateAttendance)
		admin.DELETE("/attendance/:id", handlers.AdminDeleteAttendance)
	}

	// 普通用户接口
	user := r.Group("/api/user")
	user.Use(middlewares.JWTAuthMiddleware())
	{
		user.GET("/dashboard", handlers.UserDashboard)

		// 用户可查看自己的资料
		user.GET("/profile/:id", handlers.GetPersonByID)
		user.PUT("/profile/:id", handlers.UpdatePerson)

		// 用户可查看部门信息
		user.GET("/department/:id", handlers.GetDepartmentByID)

		// 用户提交变更申请
		user.POST("/change/request", handlers.CreatePersonnel)

		// 用户可查看自己的档案
		user.GET("/profile", handlers.GetMyProfile)

		// 用户可查看自己的考勤记录
		user.POST("/attendance/checkin", handlers.UserCheckIn)
		user.POST("/attendance/checkout", handlers.UserCheckOut)
		user.GET("/attendance/my", handlers.GetMyAttendance)

		//权限矩阵(handlers\permission.go手动添加)
		user.GET("/permissions", handlers.GetPermissions)
	}

	return r
}
