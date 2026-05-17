package dao

import (
	"backend/db"
	"backend/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FetchPersonsPaged 分页 + 搜索 + JOIN 部门
func FetchPersonsPaged(page, pageSize int, keyword string) ([]models.EmployeeInfo, int64, error) {
	dbConn := db.GetDB()

	offset := (page - 1) * pageSize
	var list []models.EmployeeInfo
	var total int64

	query := dbConn.
		Table("PERSON p").
		Select(`
			p.id,
			p.emp_id,
			p.name,
			d.name AS department,
			p.job,
			p.email,
			CASE p.state 
				WHEN 1 THEN '在职'
				WHEN 0 THEN '离职'
				ELSE '未知'
			END AS status
		`).
		Joins("LEFT JOIN DEPARTMENT d ON p.dpt_id = d.id")

	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("p.name LIKE ?", kw)
	}

	err := query.
		Order("p.id DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询员工失败: %v", err)
	}

	countQuery := dbConn.Model(&models.Person{})
	if keyword != "" {
		countQuery = countQuery.Where("name LIKE ?", "%"+strings.TrimSpace(keyword)+"%")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计员工总数失败: %w", err)
	}

	return list, total, nil
}

// FetchPersonByID 查询单个员工（ID）
func FetchPersonByID(id uint) (*models.Person, error) {
	dbConn := db.GetDB()

	var p models.Person
	err := dbConn.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound(fmt.Sprintf("未找到员工 ID: %d", id))
	}
	if err != nil {
		return nil, fmt.Errorf("查询员工失败: %w", err)
	}
	return &p, nil
}

// FetchPersonByEmpID 查询单个员工（emp_id）
func FetchPersonByEmpID(empID string) (*models.Person, error) {
	dbConn := db.GetDB()

	var p models.Person
	err := dbConn.Where("emp_id = ?", empID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound(fmt.Sprintf("未找到员工 emp_id: %s", empID))
	}
	if err != nil {
		return nil, fmt.Errorf("查询员工失败: %w", err)
	}
	return &p, nil
}

// CreatePerson 创建员工（自动生成 EMP 编号）
func CreatePerson(p *models.Person) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {

		// 生成 EMP
		if p.EmpID == "" {
			var maxID int64
			if err := tx.Table("PERSON").
				Select("IFNULL(MAX(id),0)").Scan(&maxID).Error; err != nil {
				return err
			}
			p.EmpID = fmt.Sprintf("EMP%04d", maxID+1)
		}

		// 默认名称
		if p.Name == "" {
			p.Name = fmt.Sprintf("用户%s", p.EmpID)
		}

		// 插入 PERSON
		if err := tx.Create(p).Error; err != nil {
			return fmt.Errorf("插入 PERSON 失败: %v", err)
		}

		// 设置了部门 → 部门人数 +1
		if p.DptID != 0 {
			res := tx.Model(&models.Department{}).
				Where("id = ?", p.DptID).
				Update("dpt_num", gorm.Expr("dpt_num + 1"))
			if res.Error != nil {
				return fmt.Errorf("更新部门人数失败: %w", res.Error)
			}
			if res.RowsAffected == 0 {
				return NotFound(fmt.Sprintf("未找到部门 ID: %d", p.DptID))
			}
		}

		return nil
	})
}

// UpdatePerson 更新员工信息
func UpdatePerson(id uint, p models.Person) error {
	dbConn := db.GetDB()

	updates := map[string]interface{}{
		"name":       p.Name,
		"sex":        p.Sex,
		"birth":      p.Birth,
		"job":        p.Job,
		"addr":       p.Addr,
		"tel":        p.Tel,
		"email":      p.Email,
		"remark":     p.Remark,
		"updated_at": time.Now(),
	}

	res := dbConn.
		Model(&models.Person{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("更新员工失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到员工 ID: %d", id))
	}
	return nil
}

// UpdatePersonFields 按字段更新员工信息（仅更新传入字段）
func UpdatePersonFields(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	dbConn := db.GetDB()

	res := dbConn.
		Model(&models.Person{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("更新员工失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到员工 ID: %d", id))
	}
	return nil
}

// DeletePersonByEmpID 删除员工（按 emp_id）
func DeletePersonByEmpID(empID string) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {

		var p models.Person
		err := tx.Where("emp_id = ?", empID).First(&p).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound(fmt.Sprintf("未找到员工 %s", empID))
		}
		if err != nil {
			return fmt.Errorf("查询员工失败: %w", err)
		}

		// 部门人数 -1
		if p.DptID != 0 {
			res := tx.Model(&models.Department{}).
				Where("id = ?", p.DptID).
				Update("dpt_num", gorm.Expr("dpt_num - 1"))
			if res.Error != nil {
				return fmt.Errorf("更新部门人数失败: %w", res.Error)
			}
			if res.RowsAffected == 0 {
				return NotFound(fmt.Sprintf("未找到部门 ID: %d", p.DptID))
			}
		}

		res := tx.Delete(&models.Person{}, "emp_id = ?", empID)
		if res.Error != nil {
			return fmt.Errorf("删除员工失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return NotFound(fmt.Sprintf("未找到员工 %s", empID))
		}

		return nil
	})
}

// SafeDeletePerson 删除员工（按 ID）
func SafeDeletePerson(id uint) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {

		var p models.Person
		if err := tx.First(&p, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound(fmt.Sprintf("未找到员工 ID: %d", id))
			}
			return fmt.Errorf("查询员工失败: %w", err)
		}

		if p.DptID != 0 {
			res := tx.Model(&models.Department{}).
				Where("id = ?", p.DptID).
				Update("dpt_num", gorm.Expr("dpt_num - 1"))
			if res.Error != nil {
				return fmt.Errorf("更新部门人数失败: %w", res.Error)
			}
			if res.RowsAffected == 0 {
				return NotFound(fmt.Sprintf("未找到部门 ID: %d", p.DptID))
			}
		}

		res := tx.Delete(&models.Person{}, id)
		if res.Error != nil {
			return fmt.Errorf("删除员工失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return NotFound(fmt.Sprintf("未找到员工 ID: %d", id))
		}
		return nil
	})
}

// UpdatePersonDepartment 修改员工部门（按名称）
func UpdatePersonDepartment(empID string, deptName string) error {
	dbConn := db.GetDB()

	return dbConn.Transaction(func(tx *gorm.DB) error {

		var p models.Person
		if err := tx.Where("emp_id = ?", empID).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound(fmt.Sprintf("未找到员工 %s", empID))
			}
			return fmt.Errorf("查询员工失败: %w", err)
		}

		var target models.Department
		if err := tx.Where("name = ?", deptName).First(&target).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound(fmt.Sprintf("未找到部门: %s", deptName))
			}
			return fmt.Errorf("查询目标部门失败: %w", err)
		}

		if p.DptID != target.ID {
			// 原部门 -1
			if p.DptID != 0 {
				res := tx.Model(&models.Department{}).
					Where("id = ?", p.DptID).
					Update("dpt_num", gorm.Expr("dpt_num - 1"))
				if res.Error != nil {
					return fmt.Errorf("更新原部门人数失败: %w", res.Error)
				}
				if res.RowsAffected == 0 {
					return NotFound(fmt.Sprintf("未找到原部门 ID: %d", p.DptID))
				}
			}
			// 新部门 +1
			res := tx.Model(&models.Department{}).
				Where("id = ?", target.ID).
				Update("dpt_num", gorm.Expr("dpt_num + 1"))
			if res.Error != nil {
				return fmt.Errorf("更新新部门人数失败: %w", res.Error)
			}
			if res.RowsAffected == 0 {
				return NotFound(fmt.Sprintf("未找到目标部门 ID: %d", target.ID))
			}
		}

		res := tx.Model(&models.Person{}).
			Where("emp_id = ?", empID).
			Updates(map[string]interface{}{
				"dpt_id":     target.ID,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return fmt.Errorf("更新员工部门失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return NotFound(fmt.Sprintf("未找到员工 %s", empID))
		}
		return nil
	})
}

// UpdatePersonState 修改员工状态（离职/在职）
func UpdatePersonState(empID string, state int) error {
	dbConn := db.GetDB()

	res := dbConn.Model(&models.Person{}).
		Where("emp_id = ?", empID).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("更新员工状态失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到员工 %s", empID))
	}
	return nil
}

// UpdatePersonJob 修改员工职位
func UpdatePersonJob(empID string, newJob string) error {
	dbConn := db.GetDB()

	res := dbConn.Model(&models.Person{}).
		Where("emp_id = ?", empID).
		Updates(map[string]interface{}{
			"job":        newJob,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("更新员工岗位失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到员工 %s", empID))
	}
	return nil
}
