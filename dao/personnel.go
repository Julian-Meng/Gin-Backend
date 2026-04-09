package dao

import (
	"backend/db"
	"backend/models"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// CreatePersonnelChange 创建人事变更申请
func CreatePersonnelChange(p *models.Personnel) error {
	dbConn := db.GetDB()

	//校验 emp_id 必须是工号而不是纯数字ID
	if isPureNumber(p.EmpID) {
		return fmt.Errorf("emp_id '%s' 非法，请使用员工工号（如 EMP0004）", p.EmpID)
	}

	p.State = 0 // 默认待审批
	return dbConn.Create(p).Error
}

// FetchPersonnelPaged 分页查询（带员工 & 部门信息）
func FetchPersonnelPaged(limit, offset int) ([]models.Personnel, int64, error) {
	dbConn := db.GetDB()

	var list []models.Personnel
	var total int64

	// 统计总数
	dbConn.Model(&models.Personnel{}).Count(&total)

	// 主查询（JOIN）
	err := dbConn.Table("PERSONNEL p").
		Select(`
			p.id, 
			p.emp_id,
			per.name AS emp_name,
			d1.name AS current_dpt, 
			p.target_dpt,
			d2.name AS target_dpt_name,
			p.change_type,
			p.description,
			p.leave_start_at,
			p.leave_end_at,
			p.leave_reason,
			p.leave_type,
			p.handover_note,
			p.reject_reason,
			p.state,
			p.approver,
			p.created_at,
			p.approve_at
		`).
		Joins("LEFT JOIN PERSON per ON p.emp_id = per.emp_id").
		Joins("LEFT JOIN DEPARTMENT d1 ON per.dpt_id = d1.id").
		Joins("LEFT JOIN DEPARTMENT d2 ON p.target_dpt = d2.id").
		Order("p.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&list).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询变更记录失败: %v", err)
	}

	return list, total, nil
}

// GetPersonnelByID 获取单条人事变更详情
func GetPersonnelByID(id uint) (*models.Personnel, error) {
	dbConn := db.GetDB()

	var p models.Personnel

	err := dbConn.Table("PERSONNEL p").
		Select(`
			p.id, 
			p.emp_id,
			per.name AS emp_name,
			d1.name AS current_dpt, 
			p.target_dpt,
			d2.name AS target_dpt_name,
			p.change_type,
			p.description,
			p.leave_start_at,
			p.leave_end_at,
			p.leave_reason,
			p.leave_type,
			p.handover_note,
			p.reject_reason,
			p.state,
			p.approver,
			p.created_at,
			p.approve_at
		`).
		Joins("LEFT JOIN PERSON per ON p.emp_id = per.emp_id").
		Joins("LEFT JOIN DEPARTMENT d1 ON per.dpt_id = d1.id").
		Joins("LEFT JOIN DEPARTMENT d2 ON p.target_dpt = d2.id").
		Where("p.id = ?", id).
		Scan(&p).Error

	if err != nil {
		return nil, fmt.Errorf("查询变更详情失败: %v", err)
	}

	return &p, nil
}

// ApprovePersonnelChange 管理员审批（通过 / 驳回）
//
// 变更类型：
// 1 → 调部门（更新 PERSON.dpt_id + 部门人数 +/-）
// 2 → 调岗位（更新 PERSON.job）
// 3 → 离职（更新 PERSON.state=0 + 部门人数 -1）
// 4 → 请假（仅记录审批结果，不改人员主数据）
func ApprovePersonnelChange(id uint, approver string, approve bool, rejectReason string) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {

		// 1. 获取变更记录
		var record models.Personnel
		if err := tx.First(&record, id).Error; err != nil {
			return fmt.Errorf("查询变更记录失败: %v", err)
		}

		// 校验 emp_id 是否合法
		if isPureNumber(record.EmpID) {
			return fmt.Errorf("emp_id '%s' 非法，请使用员工工号（如 EMP0004）", record.EmpID)
		}

		if record.State != 0 {
			return fmt.Errorf("该申请已审批，无法重复操作")
		}

		// 2. 获取员工信息
		var person models.Person
		if err := tx.Where("emp_id = ?", record.EmpID).First(&person).Error; err != nil {
			return fmt.Errorf("员工不存在: %s", record.EmpID)
		}

		// 3. 审批逻辑
		newState := 2 // 默认驳回
		if approve {
			newState = 1 // 审批通过

			switch record.ChangeType {

			case 1: // 调部门
				oldDpt := person.DptID
				newDpt := record.TargetDpt

				// 更新员工部门
				if err := tx.Model(&models.Person{}).
					Where("emp_id = ?", record.EmpID).
					Update("dpt_id", newDpt).Error; err != nil {
					return fmt.Errorf("更新员工部门失败: %v", err)
				}

				// 目标部门人数 +1
				tx.Model(&models.Department{}).
					Where("id = ?", newDpt).
					Update("dpt_num", gorm.Expr("dpt_num + 1"))

				// 原部门人数 -1
				if oldDpt != 0 && oldDpt != newDpt {
					tx.Model(&models.Department{}).
						Where("id = ?", oldDpt).
						Update("dpt_num", gorm.Expr("dpt_num - 1"))
				}

			case 2: // 调岗位
				newJob := record.Description
				if err := tx.Model(&models.Person{}).
					Where("emp_id = ?", record.EmpID).
					Update("job", newJob).Error; err != nil {
					return fmt.Errorf("更新员工岗位失败: %v", err)
				}

			case 3: // 离职
				if err := tx.Model(&models.Person{}).
					Where("emp_id = ?", record.EmpID).
					Update("state", 0).Error; err != nil {
					return fmt.Errorf("更新员工状态失败: %v", err)
				}

				// 部门人数 -1
				tx.Model(&models.Department{}).
					Where("id = ?", person.DptID).
					Update("dpt_num", gorm.Expr("dpt_num - 1"))

			case 4: // 请假
				// 请假审批通过时仅更新审批结果，不改人员主数据
			}
		}

		// 4. 更新人事变更记录
		now := time.Now()
		if approve {
			rejectReason = ""
		}
		return tx.Model(&models.Personnel{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"state":         newState,
				"approver":      approver,
				"reject_reason": rejectReason,
				"approve_at":    &now,
			}).Error
	})
}

func isPureNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
