package models

import "time"

// Person 员工表结构（仅模型）
type Person struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	EmpID     string     `gorm:"size:32;uniqueIndex;not null" json:"emp_id"`
	Name      string     `gorm:"size:64;not null" json:"name"`
	Auth      int        `gorm:"default:0" json:"auth"`
	Sex       string     `gorm:"size:8" json:"sex"`
	Birth     *time.Time `json:"birth"`
	DptID     uint       `gorm:"index" json:"dpt_id"`
	Job       string     `gorm:"size:64" json:"job"`
	Addr      string     `gorm:"size:255" json:"addr"`
	Tel       string     `gorm:"size:32" json:"tel"`
	Email     string     `gorm:"size:64" json:"email"`
	State     int        `gorm:"default:1" json:"state"`
	Remark    string     `gorm:"size:255" json:"remark"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"update_at"`
}

func (Person) TableName() string {
	return "PERSON"
}

// 列表显示结构
type EmployeeInfo struct {
	ID         uint   `json:"id"`
	EmpID      string `json:"emp_id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Job        string `json:"job"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}
