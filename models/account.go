package models

import "time"

// Account 账号表结构
type Account struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string     `gorm:"size:255;not null" json:"password,omitempty"`
	EmpID     string     `gorm:"size:32;index;not null" json:"emp_id"`
	Role      string     `gorm:"size:16;default:'staff'" json:"role"`
	Status    int        `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"create_at"`
	LastLogin *time.Time `json:"last_login"`
}

func (Account) TableName() string {
	return "ACCOUNT"
}
