package models

import "time"

// Notice 公告表结构（仅模型）
type Notice struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Title     string     `gorm:"size:128;not null" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	Publisher string     `gorm:"size:64" json:"publisher"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

func (Notice) TableName() string {
	return "NOTICE"
}
