package models

import (
	"backend/db"
	"database/sql"
	"fmt"
	"strings"
)

// Department 部门结构体
type Department struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Manager     string         `json:"manager"`
	Intro       string         `json:"intro"`
	CreateAt    string         `json:"create_at"`
	Location    string         `json:"location"`
	DptNum      int            `json:"dpt_num"`
	FullNum     int            `json:"full_num"`
	IsFull      int            `json:"is_full"`
	MemberCount sql.NullInt64  `json:"member_count,omitempty"` // 可选字段（统计人数）
	Description sql.NullString `json:"description,omitempty"`  // 兼容前端描述字段
}

// ✅ 分页 + 搜索查询部门
func FetchDepartmentsPaged(page, pageSize int, keyword string) ([]Department, int, error) {
	offset := (page - 1) * pageSize
	where := ""
	args := []interface{}{}

	if keyword != "" {
		where = "WHERE name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(keyword)+"%")
	}

	query := fmt.Sprintf(`
    SELECT d.id, d.name, d.manager, d.intro, d.create_at, d.location,
        d.dpt_num, d.full_num, d.is_full,
        COUNT(p.id) AS member_count
    FROM DEPARTMENT d
    LEFT JOIN PERSON p ON d.id = p.dpt_id
    %s
    GROUP BY d.id
    ORDER BY d.id DESC
    LIMIT ? OFFSET ?
	`, where)

	args = append(args, pageSize, offset)
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var dpts []Department
	for rows.Next() {
		var d Department
		rows.Scan(&d.ID, &d.Name, &d.Manager, &d.Intro, &d.CreateAt,
			&d.Location, &d.DptNum, &d.FullNum, &d.IsFull, &d.MemberCount)
		dpts = append(dpts, d)
	}

	// 统计总数
	var total int
	countQuery := "SELECT COUNT(*) FROM DEPARTMENT"
	if keyword != "" {
		countQuery += " WHERE name LIKE ?"
		_ = db.DB.QueryRow(countQuery, "%"+keyword+"%").Scan(&total)
	} else {
		_ = db.DB.QueryRow(countQuery).Scan(&total)
	}

	return dpts, total, nil
}

// ✅ 新增部门
func InsertDepartment(d Department) error {
	if d.FullNum == 0 {
		d.FullNum = 20 // 默认部门上限人数，可自行调整
	}
	if d.DptNum < 0 {
		d.DptNum = 0
	}
	if d.DptNum >= d.FullNum {
		d.IsFull = 1
	} else {
		d.IsFull = 0
	}
	_, err := db.DB.Exec(`
		INSERT INTO DEPARTMENT (name, manager, intro, create_at, location, dpt_num, full_num, is_full)
		VALUES (?, ?, ?, datetime('now'), ?, ?, ?, ?)
	`, d.Name, d.Manager, d.Intro, d.Location, d.DptNum, d.FullNum, d.IsFull)
	if err != nil {
		fmt.Println("❌ 插入部门失败:", err)
	}
	return err
}

// ✅ 更新部门
func UpdateDepartment(id string, d Department) error {
	if d.DptNum >= d.FullNum {
		d.IsFull = 1
	} else {
		d.IsFull = 0
	}
	_, err := db.DB.Exec(`
		UPDATE DEPARTMENT
		SET name=?, manager=?, intro=?, location=?, dpt_num=?, full_num=?, is_full=?
		WHERE id=?
	`, d.Name, d.Manager, d.Intro, d.Location, d.DptNum, d.FullNum, d.IsFull, id)
	if err != nil {
		fmt.Println("❌ 更新部门失败:", err)
	}
	return err
}

// ✅ 删除部门
func DeleteDepartment(id string) error {
	var cnt int
	db.DB.QueryRow(`SELECT COUNT(*) FROM PERSON WHERE dpt_id = ?`, id).Scan(&cnt)
	if cnt > 0 {
		return fmt.Errorf("部门下仍有关联员工，无法删除")
	}

	_, err := db.DB.Exec(`DELETE FROM DEPARTMENT WHERE id = ?`, id)
	if err != nil {
		fmt.Println("❌ 删除部门失败:", err)
	}
	return err
}

// ✅ 按 ID 查询部门详情
func FetchDepartmentByID(id string) (Department, error) {
	var d Department
	err := db.DB.QueryRow(`
		SELECT id, name, manager, intro, create_at, location, dpt_num, full_num, is_full
		FROM DEPARTMENT
		WHERE id = ?
	`).Scan(&d.ID, &d.Name, &d.Manager, &d.Intro, &d.CreateAt, &d.Location, &d.DptNum, &d.FullNum, &d.IsFull)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return Department{}, fmt.Errorf("未找到部门 ID: %s", id)
		}
		fmt.Println("❌ 查询部门详情失败:", err)
		return Department{}, err
	}
	return d, nil
}

// ✅ 按部门 ID 调整人数（增减）
func UpdateDepartmentCount(dptID int, delta int) error {
	_, err := db.DB.Exec(`UPDATE DEPARTMENT SET dpt_num = dpt_num + ? WHERE id = ?`, delta, dptID)
	return err
}

// ✅ 按部门名称调整人数（用于目标部门）
func UpdateDepartmentCountByName(name string, delta int) error {
	_, err := db.DB.Exec(`UPDATE DEPARTMENT SET dpt_num = dpt_num + ? WHERE name = ?`, delta, name)
	return err
}

// ==========================
// 确保至少存在一个“默认部门”
// ==========================
func EnsureDefaultDepartment() error {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM DEPARTMENT`).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查部门表失败: %v", err)
	}

	// 如果没有任何部门，自动创建一个
	if count == 0 {
		_, err := db.DB.Exec(`
			INSERT INTO DEPARTMENT (name, manager, intro, location, dpt_num, full_num, is_full)
			VALUES ('未分配部门', '系统', '系统默认部门', '未知', 0, 50, 0)
		`)
		if err != nil {
			return fmt.Errorf("创建默认部门失败: %v", err)
		}
		fmt.Println("📦 已自动创建默认部门: '未分配部门'")
	}
	return nil
}
