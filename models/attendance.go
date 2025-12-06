package models

import "time"

// Attendance 考勤表结构
// 使用 emp_id 关联员工(Person / Account)，保持与现有设计一致。
type Attendance struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	EmpID     string     `gorm:"size:32;index;not null" json:"emp_id"` // 员工编号
	Date      time.Time  `gorm:"type:date;index;not null" json:"date"` // 考勤日期（只关心年月日）
	CheckIn   *time.Time `json:"check_in"`                             // 签到时间
	CheckOut  *time.Time `json:"check_out"`                            // 签退时间
	Status    int        `gorm:"default:0;not null" json:"status"`     // 0缺勤 1正常 2迟到 3早退 4迟到且早退
	Remark    string     `gorm:"size:255" json:"remark"`               // 备注（管理员手动说明）
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`     // 创建时间
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`     // 更新时间
}

// TableName 显式指定表名（与其他表保持大写风格）
func (Attendance) TableName() string {
	return "ATTENDANCE"
}

// AttendanceDetail 用于列表 / 管理员查询时的联表展示结构
// 会在 DAO 层通过 JOIN PERSON / DEPARTMENT 填充。
type AttendanceDetail struct {
	ID          uint       `json:"id"`
	EmpID       string     `json:"emp_id"`
	Name        string     `json:"name"`       // 员工姓名（来自 PERSON）
	Department  string     `json:"department"` // 部门名称（来自 DEPARTMENT）
	Job         string     `json:"job"`        // 职位（来自 PERSON）
	Date        time.Time  `json:"date"`
	CheckIn     *time.Time `json:"check_in"`
	CheckOut    *time.Time `json:"check_out"`
	Status      int        `json:"status"`       // 同上：0缺勤 1正常 2迟到 3早退 4迟到且早退
	StatusLabel string     `json:"status_label"` // 人类可读的中文状态，在 SQL 或 DAO 中组装
	Remark      string     `json:"remark"`
}
