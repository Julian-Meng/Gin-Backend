package models

import (
	"backend/db"
	"database/sql"
	"fmt"
	"log"
)

// ======================================
// Personnel 数据结构
// ======================================
type Personnel struct {
	ID            int            `json:"id"`
	EmpID         string         `json:"emp_id"`
	EmpName       string         `json:"emp_name,omitempty"`
	CurrentDpt    string         `json:"current_dpt,omitempty"`
	TargetDpt     int            `json:"target_dpt"` // ✅ 改为部门ID
	TargetDptName string         `json:"target_dpt_name,omitempty"`
	ChangeType    int            `json:"change_type"`
	Description   string         `json:"description"`
	State         int            `json:"state"`
	Approver      sql.NullString `json:"approver"`
	CreateAt      string         `json:"create_at"`
	ApproveAt     sql.NullString `json:"approve_at"`
}

// ======================================
// 创建人事变更申请
// ======================================
func CreatePersonnelChange(p Personnel) error {
	_, err := db.DB.Exec(`
		INSERT INTO PERSONNEL (emp_id, change_type, description, state, target_dpt, approver, create_at)
		VALUES (?, ?, ?, 0, ?, '', datetime('now'))
	`, p.EmpID, p.ChangeType, p.Description, p.TargetDpt)
	if err != nil {
		log.Printf("❌ 新增变更记录失败: %v", err)
		return err
	}
	log.Printf("✅ 已创建人事变更申请: emp_id=%s target_dpt=%d", p.EmpID, p.TargetDpt)
	return nil
}

// ======================================
// 分页查询所有人事变更（带详情）
// ======================================
func FetchPersonnelPaged(limit, offset int) ([]Personnel, int, error) {
	rows, err := db.DB.Query(`
		SELECT 
			p.id, p.emp_id, per.name AS emp_name, d1.name AS current_dpt, 
			p.target_dpt, d2.name AS target_dpt_name, 
			p.change_type, p.description, p.state, 
			p.approver, p.create_at, p.approve_at
		FROM PERSONNEL p
		LEFT JOIN PERSON per ON p.emp_id = per.emp_id
		LEFT JOIN DEPARTMENT d1 ON per.dpt_id = d1.id
		LEFT JOIN DEPARTMENT d2 ON p.target_dpt = d2.id
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?;
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询变更记录失败: %v", err)
	}
	defer rows.Close()

	var list []Personnel
	for rows.Next() {
		var p Personnel
		err := rows.Scan(
			&p.ID, &p.EmpID, &p.EmpName, &p.CurrentDpt,
			&p.TargetDpt, &p.TargetDptName, &p.ChangeType,
			&p.Description, &p.State, &p.Approver,
			&p.CreateAt, &p.ApproveAt,
		)
		if err != nil {
			log.Printf("⚠️ 扫描行出错: %v", err)
			continue
		}
		if !p.Approver.Valid {
			p.Approver.String = ""
		}
		if !p.ApproveAt.Valid {
			p.ApproveAt.String = ""
		}
		list = append(list, p)
	}

	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM PERSONNEL`).Scan(&total)
	return list, total, nil
}

// ======================================
// 管理员审批变更
// ======================================
func ApprovePersonnelChange(id int, approver string, approve bool) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 查询当前状态
	var currentState int
	err = tx.QueryRow(`SELECT state FROM PERSONNEL WHERE id = ?`, id).Scan(&currentState)
	if err != nil {
		return fmt.Errorf("查询变更状态失败: %v", err)
	}
	if currentState != 0 {
		return fmt.Errorf("该申请已审批，无法重复操作")
	}

	// 获取变更记录详情
	var empID string
	var targetDpt, changeType int
	err = tx.QueryRow(`SELECT emp_id, target_dpt, change_type FROM PERSONNEL WHERE id = ?`, id).
		Scan(&empID, &targetDpt, &changeType)
	if err != nil {
		return fmt.Errorf("查询变更信息失败: %v", err)
	}

	newState := 2 // 默认驳回
	if approve {
		newState = 1 // 审批通过

		switch changeType {
		case 1: // 调部门
			var oldDpt int
			_ = tx.QueryRow(`SELECT dpt_id FROM PERSON WHERE emp_id = ?`, empID).Scan(&oldDpt)

			// 更新员工部门
			_, err = tx.Exec(`UPDATE PERSON SET dpt_id = ?, update_at = datetime('now') WHERE emp_id = ?`,
				targetDpt, empID)
			if err != nil {
				return fmt.Errorf("更新员工部门失败: %v", err)
			}

			// 更新部门人数（目标+1，原部门-1）
			_, _ = tx.Exec(`UPDATE DEPARTMENT SET dpt_num = dpt_num + 1 WHERE id = ?`, targetDpt)
			if oldDpt != targetDpt {
				_, _ = tx.Exec(`UPDATE DEPARTMENT SET dpt_num = dpt_num - 1 WHERE id = ? AND dpt_num > 0`, oldDpt)
			}

		case 2: // 调岗位
			// 你可以在 PERSONNEL 表 description 里写新岗位
			var newJob string
			_ = tx.QueryRow(`SELECT description FROM PERSONNEL WHERE id = ?`, id).Scan(&newJob)
			_, err = tx.Exec(`UPDATE PERSON SET job = ?, update_at = datetime('now') WHERE emp_id = ?`, newJob, empID)
			if err != nil {
				return fmt.Errorf("更新员工岗位失败: %v", err)
			}

		case 3: // 离职
			_, err = tx.Exec(`UPDATE PERSON SET state = 0, update_at = datetime('now') WHERE emp_id = ?`, empID)
			if err != nil {
				return fmt.Errorf("更新员工状态失败: %v", err)
			}
			_, _ = tx.Exec(`
				UPDATE DEPARTMENT SET dpt_num = dpt_num - 1 
				WHERE id = (SELECT dpt_id FROM PERSON WHERE emp_id = ?) AND dpt_num > 0
			`, empID)
		}
	}

	// 更新人事变更记录
	_, err = tx.Exec(`
		UPDATE PERSONNEL
		SET state = ?, approver = ?, approve_at = datetime('now')
		WHERE id = ?
	`, newState, approver, id)
	if err != nil {
		return fmt.Errorf("更新变更状态失败: %v", err)
	}

	log.Printf("✅ 人事变更审批完成: id=%d 状态=%d", id, newState)
	return nil
}

// ======================================
// 获取单个人事变更详情
// ======================================
func GetPersonnelByID(id int) (Personnel, error) {
	var p Personnel
	err := db.DB.QueryRow(`
		SELECT 
			p.id, p.emp_id, per.name AS emp_name, d1.name AS current_dpt, 
			p.target_dpt, d2.name AS target_dpt_name, 
			p.change_type, p.description, p.state, 
			p.approver, p.create_at, p.approve_at
		FROM PERSONNEL p
		LEFT JOIN PERSON per ON p.emp_id = per.emp_id
		LEFT JOIN DEPARTMENT d1 ON per.dpt_id = d1.id
		LEFT JOIN DEPARTMENT d2 ON p.target_dpt = d2.id
		WHERE p.id = ?;
	`, id).Scan(
		&p.ID, &p.EmpID, &p.EmpName, &p.CurrentDpt,
		&p.TargetDpt, &p.TargetDptName, &p.ChangeType,
		&p.Description, &p.State, &p.Approver,
		&p.CreateAt, &p.ApproveAt,
	)
	if err != nil {
		return p, fmt.Errorf("查询变更详情失败: %v", err)
	}

	if !p.Approver.Valid {
		p.Approver.String = ""
	}
	if !p.ApproveAt.Valid {
		p.ApproveAt.String = ""
	}

	return p, nil
}
