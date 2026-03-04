package dao

import (
	"backend/db"
	"backend/models"
)

// GetAdminDashboardData 管理员仪表盘数据
func GetAdminDashboardData() models.AdminDashboardData {
	dbConn := db.GetDB()
	var data models.AdminDashboardData

	// 在职员工统计
	{
		var cnt int64
		dbConn.Model(&models.Person{}).
			Where("state = ?", 1).
			Count(&cnt)
		data.EmployeeCount = int(cnt)
	}

	// 部门数
	{
		var cnt int64
		dbConn.Model(&models.Department{}).Count(&cnt)
		data.DepartmentCount = int(cnt)
	}

	// 待审批人事变更数
	{
		var cnt int64
		dbConn.Model(&models.Personnel{}).
			Where("state = ?", 0).
			Count(&cnt)
		data.PendingChanges = int(cnt)
	}

	// 最近 5 条公告
	var notices []models.Notice
	dbConn.Order("id DESC").
		Limit(5).
		Find(&notices)
	data.RecentNotices = notices

	return data
}

// GetUserDashboardData 普通用户仪表盘数据
func GetUserDashboardData() models.UserDashboardData {
	dbConn := db.GetDB()
	var data models.UserDashboardData

	// 在职员工数
	{
		var cnt int64
		dbConn.Model(&models.Person{}).
			Where("state = ?", 1).
			Count(&cnt)
		data.EmployeeCount = int(cnt)
	}

	// 待审批人事变更数
	{
		var cnt int64
		dbConn.Model(&models.Personnel{}).
			Where("state = ?", 0).
			Count(&cnt)
		data.PendingChanges = int(cnt)
	}

	// 最近公告
	var notices []models.Notice
	dbConn.Order("id DESC").
		Limit(5).
		Find(&notices)
	data.RecentNotices = notices

	return data
}
