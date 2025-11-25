package models

// AdminDashboardData 管理员仪表盘数据结构
type AdminDashboardData struct {
	EmployeeCount   int      `json:"employee_count"`
	DepartmentCount int      `json:"department_count"`
	PendingChanges  int      `json:"pending_changes"`
	RecentNotices   []Notice `json:"recent_notices"`
}

// UserDashboardData 普通用户仪表盘数据结构
type UserDashboardData struct {
	EmployeeCount  int      `json:"employee_count"`
	PendingChanges int      `json:"pending_changes"`
	RecentNotices  []Notice `json:"recent_notices"`
}
