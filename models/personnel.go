package models

import "time"

// Personnel 人事变动记录（仅模型）
type Personnel struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	EmpID         string     `gorm:"size:32;not null;index" json:"emp_id"`
	EmpName       string     `json:"emp_name,omitempty" gorm:"-"`        // 关联 PERSON.name
	CurrentDpt    string     `json:"current_dpt,omitempty" gorm:"-"`     // 关联当前部门名
	TargetDpt     uint       `json:"target_dpt"`                         // 目标部门 ID
	TargetDptName string     `json:"target_dpt_name,omitempty" gorm:"-"` // 目标部门名
	ChangeType    int        `json:"change_type"`                        // 1:调部门 2:调岗 3:离职 4:请假
	Description   string     `gorm:"size:255" json:"description"`
	LeaveStartAt  *time.Time `json:"leave_start_at"`
	LeaveEndAt    *time.Time `json:"leave_end_at"`
	LeaveReason   string     `gorm:"size:500" json:"leave_reason"`
	LeaveType     string     `gorm:"size:32" json:"leave_type"`
	HandoverNote  string     `gorm:"size:255" json:"handover_note"`
	RejectReason  string     `gorm:"size:255" json:"reject_reason"`
	State         int        `gorm:"default:0" json:"state"` // 0 待审批 1 通过 2 驳回
	Approver      *string    `json:"approver"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"create_at"`
	ApproveAt     *time.Time `json:"approve_at"`
}

func (Personnel) TableName() string {
	return "PERSONNEL"
}
