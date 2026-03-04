package models

// Department 部门表结构（仅模型）
type Department struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Manager   string `gorm:"size:64" json:"manager"`
	Intro     string `gorm:"size:255" json:"intro"`
	Location  string `gorm:"size:128" json:"location"`
	DptNum    int    `gorm:"default:0" json:"dpt_num"`
	FullNum   int    `gorm:"default:0" json:"full_num"`
	IsFull    int    `gorm:"default:0" json:"is_full"`
	CreatedAt string `json:"create_at"` // 原字段名兼容旧前端
}

func (Department) TableName() string {
	return "DEPARTMENT"
}

// DepartmentWithCount 用于列表返回的结构
type DepartmentWithCount struct {
	Department
	MemberCount int    `json:"member_count"`
	Description string `json:"description,omitempty"`
}
