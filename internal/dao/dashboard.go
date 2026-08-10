package dao

import (
	"backend/internal/db"
	"backend/internal/model"
	"fmt"
)

// GetAdminDashboardData 管理员仪表盘数据
func GetAdminDashboardData() (model.AdminDashboardData, error) {
	dbConn := db.GetDB()
	var data model.AdminDashboardData

	// 在职员工统计
	{
		var cnt int64
		if err := dbConn.Model(&model.Person{}).
			Where("state = ?", 1).
			Count(&cnt).Error; err != nil {
			return model.AdminDashboardData{}, fmt.Errorf("统计在职员工失败: %w", err)
		}
		data.EmployeeCount = int(cnt)
	}

	// 部门数
	{
		var cnt int64
		if err := dbConn.Model(&model.Department{}).Count(&cnt).Error; err != nil {
			return model.AdminDashboardData{}, fmt.Errorf("统计部门数量失败: %w", err)
		}
		data.DepartmentCount = int(cnt)
	}

	// 待审批人事变更数
	{
		var cnt int64
		if err := dbConn.Model(&model.Personnel{}).
			Where("state = ?", 0).
			Count(&cnt).Error; err != nil {
			return model.AdminDashboardData{}, fmt.Errorf("统计待审批变更失败: %w", err)
		}
		data.PendingChanges = int(cnt)
	}

	// 最近 5 条公告
	var notices []model.Notice
	if err := dbConn.Order("id DESC").
		Limit(5).
		Find(&notices).Error; err != nil {
		return model.AdminDashboardData{}, fmt.Errorf("查询近期公告失败: %w", err)
	}
	data.RecentNotices = notices

	return data, nil
}

// GetUserDashboardData 普通用户仪表盘数据
func GetUserDashboardData() (model.UserDashboardData, error) {
	dbConn := db.GetDB()
	var data model.UserDashboardData

	// 在职员工数
	{
		var cnt int64
		if err := dbConn.Model(&model.Person{}).
			Where("state = ?", 1).
			Count(&cnt).Error; err != nil {
			return model.UserDashboardData{}, fmt.Errorf("统计在职员工失败: %w", err)
		}
		data.EmployeeCount = int(cnt)
	}

	// 待审批人事变更数
	{
		var cnt int64
		if err := dbConn.Model(&model.Personnel{}).
			Where("state = ?", 0).
			Count(&cnt).Error; err != nil {
			return model.UserDashboardData{}, fmt.Errorf("统计待审批变更失败: %w", err)
		}
		data.PendingChanges = int(cnt)
	}

	// 最近公告
	var notices []model.Notice
	if err := dbConn.Order("id DESC").
		Limit(5).
		Find(&notices).Error; err != nil {
		return model.UserDashboardData{}, fmt.Errorf("查询近期公告失败: %w", err)
	}
	data.RecentNotices = notices

	return data, nil
}
