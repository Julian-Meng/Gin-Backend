package dao

import (
	"backend/db"
	"backend/models"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// FetchDepartmentsPaged 分页 + 搜索 + 成员数量统计
func FetchDepartmentsPaged(page, pageSize int, keyword string) ([]models.DepartmentWithCount, int64, error) {
	dbConn := db.GetDB()

	offset := (page - 1) * pageSize
	var list []models.DepartmentWithCount
	var total int64

	query := dbConn.
		Table("DEPARTMENT d").
		Select(`
			d.id, d.name, d.manager, d.intro, d.created_at, d.location,
			d.dpt_num, d.full_num, d.is_full,
			COUNT(p.id) AS member_count
		`).
		Joins("LEFT JOIN PERSON p ON d.id = p.dpt_id")

	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("d.name LIKE ?", kw)
	}

	err := query.Group("d.id").
		Order("d.id DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&list).Error

	if err != nil {
		return nil, 0, err
	}

	// 统计总数
	countQuery := dbConn.Model(&models.Department{})
	if keyword != "" {
		countQuery = countQuery.Where("name LIKE ?", "%"+strings.TrimSpace(keyword)+"%")
	}
	countQuery.Count(&total)

	return list, total, nil
}

// InsertDepartment 新增部门
func InsertDepartment(d models.Department) error {
	dbConn := db.GetDB()

	if d.FullNum == 0 {
		d.FullNum = 20
	}
	if d.DptNum < 0 {
		d.DptNum = 0
	}

	if d.DptNum >= d.FullNum {
		d.IsFull = 1
	} else {
		d.IsFull = 0
	}

	return dbConn.Create(&d).Error
}

// UpdateDepartment 更新部门
func UpdateDepartment(id uint, d models.Department) error {
	dbConn := db.GetDB()

	if d.DptNum >= d.FullNum {
		d.IsFull = 1
	} else {
		d.IsFull = 0
	}

	updates := map[string]interface{}{
		"name":     d.Name,
		"manager":  d.Manager,
		"intro":    d.Intro,
		"location": d.Location,
		"dpt_num":  d.DptNum,
		"full_num": d.FullNum,
		"is_full":  d.IsFull,
	}

	return dbConn.
		Model(&models.Department{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteDepartment 删除部门（需检查人数）
func DeleteDepartment(id uint) error {
	dbConn := db.GetDB()

	// 检查部门下是否有人
	var count int64
	dbConn.Model(&models.Person{}).
		Where("dpt_id = ?", id).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("部门下仍有关联员工，无法删除")
	}

	return dbConn.Delete(&models.Department{}, id).Error
}

// FetchDepartmentByID 查询单个部门
func FetchDepartmentByID(id uint) (*models.Department, error) {
	dbConn := db.GetDB()

	var d models.Department
	err := dbConn.First(&d, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未找到部门 ID: %d", id)
		}
		return nil, err
	}

	return &d, nil
}

// UpdateDepartmentCount 部门人数增减（按 ID）
func UpdateDepartmentCount(dptID uint, delta int) error {
	dbConn := db.GetDB()

	return dbConn.Model(&models.Department{}).
		Where("id = ?", dptID).
		Update("dpt_num", gorm.Expr("dpt_num + ?", delta)).
		Error
}

// UpdateDepartmentCountByName 部门人数增减（按名称）
func UpdateDepartmentCountByName(name string, delta int) error {
	dbConn := db.GetDB()

	return dbConn.Model(&models.Department{}).
		Where("name = ?", name).
		Update("dpt_num", gorm.Expr("dpt_num + ?", delta)).
		Error
}
