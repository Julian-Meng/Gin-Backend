package models

import (
	"backend/db"
	"log"
)

// ==========================
// 仪表盘数据结构
// ==========================
type AdminDashboardData struct {
	EmployeeCount   int      `json:"employee_count"`   // 在职员工数
	DepartmentCount int      `json:"department_count"` // 部门总数
	PendingChanges  int      `json:"pending_changes"`  // 待审批变更数
	RecentNotices   []Notice `json:"recent_notices"`   // 最新公告
}

type UserDashboardData struct {
	EmployeeCount  int      `json:"employee_count"`  // 在职员工数
	PendingChanges int      `json:"pending_changes"` // 待审批变更
	RecentNotices  []Notice `json:"recent_notices"`  // 最新公告
}

// ==========================
// 管理员仪表盘数据
// ==========================
func GetAdminDashboardData() AdminDashboardData {
	var data AdminDashboardData

	// 在职员工数
	db.DB.QueryRow("SELECT COUNT(*) FROM PERSON WHERE state = 1").Scan(&data.EmployeeCount)

	// 部门总数
	db.DB.QueryRow("SELECT COUNT(*) FROM DEPARTMENT").Scan(&data.DepartmentCount)

	// 待审批人事变更数
	db.DB.QueryRow("SELECT COUNT(*) FROM PERSONNEL WHERE state = 0").Scan(&data.PendingChanges)

	// 最新公告（取 5 条）
	rows, err := db.DB.Query(`
		SELECT id, title, content, publisher, create_at, update_at
		FROM NOTICE
		ORDER BY id DESC
		LIMIT 5
	`)
	if err != nil {
		log.Println("⚠️ 查询公告失败:", err)
		return data
	}
	defer rows.Close()

	for rows.Next() {
		var n Notice
		err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Publisher, &n.CreateAt, &n.UpdateAt)
		if err != nil {
			log.Println("⚠️ 扫描公告失败:", err)
			continue
		}
		data.RecentNotices = append(data.RecentNotices, n)
	}

	return data
}

// ==========================
// 普通用户仪表盘数据
// ==========================
func GetUserDashboardData() UserDashboardData {
	var data UserDashboardData

	// 在职员工数
	db.DB.QueryRow("SELECT COUNT(*) FROM PERSON WHERE state = 1").Scan(&data.EmployeeCount)

	// 待审批人事变更
	db.DB.QueryRow("SELECT COUNT(*) FROM PERSONNEL WHERE state = 0").Scan(&data.PendingChanges)

	// 最近公告（取 5 条）
	rows, err := db.DB.Query(`
		SELECT id, title, content, publisher, create_at, update_at
		FROM NOTICE
		ORDER BY id DESC
		LIMIT 5
	`)
	if err != nil {
		log.Println("⚠️ 查询公告失败:", err)
		return data
	}
	defer rows.Close()

	for rows.Next() {
		var n Notice
		err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Publisher, &n.CreateAt, &n.UpdateAt)
		if err != nil {
			log.Println("⚠️ 扫描公告失败:", err)
			continue
		}
		data.RecentNotices = append(data.RecentNotices, n)
	}

	return data
}
