package dao

import (
	"backend/db"
	"backend/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// =========================
// 工具：获取“当天日期”（去掉时分秒）
// =========================
func todayDate() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0,
		now.Location(),
	)
}

// =========================
// 员工签到
// - 若当天已有签到，则返回错误
// - 若已有记录但缺少签到时间，则补上签到时间
// =========================
func CheckIn(empID string) error {
	dbConn := db.GetDB()
	today := todayDate()
	now := time.Now()

	return dbConn.Transaction(func(tx *gorm.DB) error {
		var att models.Attendance

		err := tx.
			Where("emp_id = ? AND date = ?", empID, today).
			First(&att).Error

		if err == gorm.ErrRecordNotFound {
			// 当天没有记录 → 新建一条
			newAtt := models.Attendance{
				EmpID:   empID,
				Date:    today,
				CheckIn: &now,
				Status:  1, // 先默认正常，迟到逻辑后续可以再加
			}
			if err := tx.Create(&newAtt).Error; err != nil {
				return fmt.Errorf("创建签到记录失败: %v", err)
			}
			return nil
		}

		if err != nil {
			return fmt.Errorf("查询签到记录失败: %v", err)
		}

		// 已存在当天记录
		if att.CheckIn != nil {
			return fmt.Errorf("今日已签到，无需重复签到")
		}

		// 只补签到时间
		return tx.Model(&models.Attendance{}).
			Where("id = ?", att.ID).
			Updates(map[string]interface{}{
				"check_in":   now,
				"updated_at": time.Now(),
			}).Error
	})
}

// =========================
// 员工签退
// - 若当天记录不存在，可以选择报错
// - 若存在但已签退，则返回错误
// =========================
func CheckOut(empID string) error {
	dbConn := db.GetDB()
	today := todayDate()
	now := time.Now()

	return dbConn.Transaction(func(tx *gorm.DB) error {
		var att models.Attendance

		err := tx.
			Where("emp_id = ? AND date = ?", empID, today).
			First(&att).Error

		if err == gorm.ErrRecordNotFound {
			// 没有签到记录也尝试签退 → 可以选择自动建一条，也可以直接报错
			return fmt.Errorf("今日未找到考勤记录，请先签到或联系管理员处理")
		}

		if err != nil {
			return fmt.Errorf("查询签退记录失败: %v", err)
		}

		if att.CheckOut != nil {
			return fmt.Errorf("今日已签退，无需重复操作")
		}

		// 更新签退时间
		return tx.Model(&models.Attendance{}).
			Where("id = ?", att.ID).
			Updates(map[string]interface{}{
				"check_out":  now,
				"updated_at": time.Now(),
			}).Error
	})
}

// =========================
// 员工个人考勤查询（分页）
// - 按 emp_id + 日期区间
// - 按日期倒序
// =========================
func GetAttendanceByEmpID(empID string, startDate, endDate time.Time, page, pageSize int) ([]models.Attendance, int64, error) {
	dbConn := db.GetDB()

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 规范化 start/end，只保留日期
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), endDate.Location())

	var list []models.Attendance
	var total int64

	query := dbConn.Model(&models.Attendance{}).
		Where("emp_id = ?", empID).
		Where("date >= ? AND date <= ?", start, end)

	// 总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计考勤记录失败: %v", err)
	}

	// 分页数据
	if err := query.
		Order("date DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询考勤记录失败: %v", err)
	}

	return list, total, nil
}

// =========================
// 管理员：按条件搜索考勤记录（分页）
// 支持：emp_id、dpt_id、日期区间
// 返回联表后的 AttendanceDetail
// =========================
func SearchAttendance(empID string, dptID uint, startDate, endDate time.Time, page, pageSize int) ([]models.AttendanceDetail, int64, error) {
	dbConn := db.GetDB()

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), endDate.Location())

	var list []models.AttendanceDetail
	var total int64

	// 主查询：联表 PERSON / DEPARTMENT
	query := dbConn.
		Table("ATTENDANCE a").
		Select(`
			a.id,
			a.emp_id,
			p.name AS name,
			d.name AS department,
			p.job AS job,
			a.date,
			a.check_in,
			a.check_out,
			a.status,
			CASE a.status
				WHEN 0 THEN '缺勤'
				WHEN 1 THEN '正常'
				WHEN 2 THEN '迟到'
				WHEN 3 THEN '早退'
				WHEN 4 THEN '迟到且早退'
				ELSE '未知'
			END AS status_label,
			a.remark
		`).
		Joins("LEFT JOIN PERSON p ON a.emp_id = p.emp_id").
		Joins("LEFT JOIN DEPARTMENT d ON p.dpt_id = d.id").
		Where("a.date >= ? AND a.date <= ?", start, end)

	// 可选过滤条件
	if empID != "" {
		query = query.Where("a.emp_id = ?", empID)
	}
	if dptID != 0 {
		query = query.Where("p.dpt_id = ?", dptID)
	}

	// 统计总数
	var count int64
	countQuery := dbConn.Table("ATTENDANCE a").
		Joins("LEFT JOIN PERSON p ON a.emp_id = p.emp_id").
		Where("a.date >= ? AND a.date <= ?", start, end)
	if empID != "" {
		countQuery = countQuery.Where("a.emp_id = ?", empID)
	}
	if dptID != 0 {
		countQuery = countQuery.Where("p.dpt_id = ?", dptID)
	}
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("统计考勤总数失败: %v", err)
	}
	total = count

	// 查询分页结果
	if err := query.
		Order("a.date DESC, a.emp_id ASC").
		Limit(pageSize).
		Offset(offset).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询考勤列表失败: %v", err)
	}

	return list, total, nil
}

// =========================
// 管理员：更新考勤记录（如状态 / 备注 / 时间）
// 仅更新调用方提供的字段
// =========================
func UpdateAttendance(id uint, updates map[string]interface{}) error {
	dbConn := db.GetDB()

	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}
	updates["updated_at"] = time.Now()

	return dbConn.Model(&models.Attendance{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// =========================
// 管理员：删除考勤记录
// =========================
func DeleteAttendance(id uint) error {
	dbConn := db.GetDB()

	return dbConn.Delete(&models.Attendance{}, id).Error
}
