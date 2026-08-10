package main

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/middleware/errorx"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(errorx.RequestID(), gin.Logger(), errorx.Recovery())

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-Refresh-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
	r.StaticFile("/bt", "./static/backend_test.html")
	r.StaticFile("/jv", "./static/json_viewer.html")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 无需登录
	r.GET("/api/captcha", handler.GetCaptcha)
	r.POST("/api/login", handler.Login)
	r.POST("/api/register", handler.Register)
	r.GET("/api/notice", handler.GetAllNotices)
	r.POST("/api/chat", handler.ChatWithAI)
	r.GET("/api/chat/ws", handler.ChatWS)

	// 管理员接口
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminOnly())
	{
		// 仪表盘
		admin.GET("/dashboard", handler.AdminDashboard)

		// 管理员组AI分析
		admin.GET("/ai/analyze/dashboard", handler.AdminAnalyzeDashboardAI)

		// 员工管理（person.go）
		admin.GET("/persons", handler.GetPersons)
		admin.POST("/person", handler.CreatePerson)
		admin.PUT("/person/:id", handler.UpdatePerson)
		admin.DELETE("/person/:id", handler.DeletePersonByID)
		admin.DELETE("/person/emp/:emp_id", handler.DeletePersonByEmpID)
		admin.GET("/person/:id", handler.GetPersonByID)
		admin.PUT("/person/job", handler.ChangePersonJob)
		admin.PUT("/person/state", handler.ChangePersonState)
		admin.PUT("/person/change-dept", handler.ChangePersonDepartment)

		// 部门管理（department.go）
		admin.GET("/departments", handler.GetDepartments)
		admin.GET("/department/:id", handler.GetDepartmentByID)
		admin.POST("/department", handler.CreateDepartment)
		admin.PUT("/department/:id", handler.UpdateDepartment)
		admin.DELETE("/department/:id", handler.DeleteDepartment)

		// 人事变更（personnel.go）
		admin.GET("/changes", handler.GetPersonnelList)
		admin.GET("/change/:id", handler.GetPersonnelByID)
		admin.POST("/change", handler.CreatePersonnel)
		admin.PUT("/change/approve", handler.ApprovePersonnel)

		// 账号管理（account.go）
		admin.GET("/accounts", handler.GetAllAccounts)
		admin.POST("/account", handler.CreateAccount)
		admin.PUT("/account/:id", handler.UpdateAccount)
		admin.DELETE("/account/:id", handler.DeleteAccount)

		// 公告管理（notice.go）
		admin.POST("/notice", handler.CreateNotice)
		admin.PUT("/notice/:id", handler.UpdateNotice)
		admin.DELETE("/notice/:id", handler.DeleteNotice)
		admin.GET("/notice/:id", handler.GetNoticeByID)

		// 个人档案（profile.go）
		admin.GET("/person/profile/:emp_id", handler.GetPersonProfile)

		// 考勤管理（attendance.go）
		admin.GET("/attendance", handler.AdminSearchAttendance)
		admin.PUT("/attendance/:id", handler.AdminUpdateAttendance)
		admin.DELETE("/attendance/:id", handler.AdminDeleteAttendance)
	}

	// 普通用户接口
	user := r.Group("/api/user")
	user.Use(middleware.JWTAuthMiddleware())
	{
		user.GET("/dashboard", handler.UserDashboard)

		// 用户组AI分析
		user.GET("/ai/analyze/dashboard", handler.UserAnalyzeDashboardAI)

		// 用户可查看自己的资料
		user.GET("/profile/:id", handler.GetPersonByID)
		user.PUT("/profile/:id", handler.UpdatePerson)

		// 用户可查看部门信息
		user.GET("/department/:id", handler.GetDepartmentByID)

		// 用户提交变更申请
		user.POST("/change/request", handler.CreatePersonnel)
		user.GET("/changes", handler.GetMyPersonnelList)

		// 用户可查看自己的档案
		user.GET("/profile", handler.GetMyProfile)
		user.PUT("/profile", handler.UpdateMyProfile)

		// 用户可查看自己的考勤记录
		user.POST("/attendance/checkin", handler.UserCheckIn)
		user.POST("/attendance/checkout", handler.UserCheckOut)
		user.GET("/attendance/my", handler.GetMyAttendance)

		//权限矩阵(internal/handler/permission.go手动添加)
		user.GET("/permissions", handler.GetPermissions)
	}

	// 聊天接口（登录后）
	chat := r.Group("/api/chat")
	chat.Use(middleware.JWTAuthMiddleware())
	{
		chat.GET("/messages/:id", handler.GetChatMessages)

		chat.POST("/user/message", handler.UserSendChatMessage)
		chat.GET("/user/sessions", handler.UserListChatSessions)

		chat.POST("/admin/message", handler.AdminSendChatMessage)
		chat.GET("/admin/sessions", handler.AdminListChatSessions)
		chat.POST("/admin/sessions/claim", handler.AdminClaimWaitingSessions)
		chat.POST("/admin/sessions/:id/claim", handler.AdminClaimSessionByID)
	}

	return r
}
