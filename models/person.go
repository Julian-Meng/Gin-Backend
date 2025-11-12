package models

import (
	"backend/db"
	"fmt"
)

// =========================
// Person 基础结构
// =========================
type Person struct {
	ID       int    `json:"id"`
	EmpID    string `json:"emp_id"`
	Name     string `json:"name"`
	Auth     int    `json:"auth"`
	Sex      string `json:"sex"`
	Birth    string `json:"birth"`
	DptID    int    `json:"dpt_id"`
	Job      string `json:"job"`
	Addr     string `json:"addr"`
	Tel      string `json:"tel"`
	Email    string `json:"email"`
	State    int    `json:"state"`
	Remark   string `json:"remark"`
	CreateAt string `json:"create_at"`
	UpdateAt string `json:"update_at"`
}

// =========================
// 列表分页结构
// =========================
type EmployeeInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Job        string `json:"job"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	EmpID      string `json:"emp_id"`
}

// =========================
// 分页查询 + 模糊搜索 + 部门名
// =========================
func FetchPersonsPaged(page, pageSize int, keyword string) ([]EmployeeInfo, int, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT 
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
		FROM PERSON p
		LEFT JOIN department d ON p.dpt_id = d.id
		WHERE p.name LIKE ?
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.DB.Query(query, "%"+keyword+"%", pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询员工失败: %v", err)
	}
	defer rows.Close()

	var list []EmployeeInfo
	for rows.Next() {
		var e EmployeeInfo
		rows.Scan(&e.ID, &e.EmpID, &e.Name, &e.Department, &e.Job, &e.Email, &e.Status)
		list = append(list, e)
	}

	// 查询总数
	var total int
	err = db.DB.QueryRow(`SELECT COUNT(*) FROM PERSON WHERE name LIKE ?`, "%"+keyword+"%").Scan(&total)
	if err != nil {
		total = len(list)
	}
	return list, total, nil
}

// =========================
// 查询单个员工（详情）
// =========================
func FetchPersonByID(id string) (Person, error) {
	row := db.DB.QueryRow(`
		SELECT id, emp_id, name, auth, sex, birth, dpt_id, job, addr, tel, email, state, remark, create_at, update_at
		FROM PERSON WHERE id = ?
	`, id)

	var p Person
	err := row.Scan(&p.ID, &p.EmpID, &p.Name, &p.Auth, &p.Sex, &p.Birth,
		&p.DptID, &p.Job, &p.Addr, &p.Tel, &p.Email, &p.State, &p.Remark,
		&p.CreateAt, &p.UpdateAt)

	return p, err
}

// =========================
// 更新员工信息
// =========================
// =========================
// 更新员工信息（字段白名单）
// =========================
func UpdatePerson(id string, p Person) error {
	// ⚙️ 仅允许修改部分字段
	query := `
		UPDATE PERSON
		SET 
			name = ?, 
			sex = ?, 
			birth = ?, 
			job = ?, 
			addr = ?, 
			tel = ?, 
			email = ?, 
			remark = ?, 
			update_at = datetime('now')
		WHERE id = ?
	`
	_, err := db.DB.Exec(query,
		p.Name, p.Sex, p.Birth, p.Job,
		p.Addr, p.Tel, p.Email, p.Remark, id)
	return err
}

// =========================
// 新增员工
// =========================
func CreatePerson(p Person) error {
	// 若前端未传 emp_id，则自动生成
	if p.EmpID == "" {
		var nextID int
		err := db.DB.QueryRow(`SELECT IFNULL(MAX(id), 0) + 1 FROM PERSON`).Scan(&nextID)
		if err != nil {
			return fmt.Errorf("生成 emp_id 失败: %v", err)
		}
		p.EmpID = fmt.Sprintf("EMP%04d", nextID)
	}

	// 确保 name 不为空
	if p.Name == "" {
		p.Name = fmt.Sprintf("用户%s", p.EmpID)
	}

	_, err := db.DB.Exec(`
		INSERT INTO person 
		(emp_id, name, auth, sex, birth, dpt_id, job, addr, tel, email, state, remark, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, p.EmpID, p.Name, p.Auth, p.Sex, p.Birth, p.DptID, p.Job, p.Addr, p.Tel, p.Email, p.State, p.Remark)

	// if err != nil {
	// 	fmt.Println("❌ 插入 PERSON 失败:", err)
	// }
	return err
}

// =========================
// 删除员工
// =========================
func DeletePerson(empID string) error {
	_, err := db.DB.Exec(`DELETE FROM PERSON WHERE emp_id = ?`, empID)
	return err
}

func SafeDeletePerson(id string) error {
	tx, _ := db.DB.Begin()
	// 查询部门 ID
	var dptID int
	_ = tx.QueryRow(`SELECT dpt_id FROM PERSON WHERE id = ?`, id).Scan(&dptID)
	tx.Exec(`UPDATE DEPARTMENT SET dpt_num = dpt_num - 1 WHERE id = ?`, dptID)
	tx.Exec(`DELETE FROM PERSON WHERE id = ?`, id)
	return tx.Commit()
}

// 根据员工编号查找员工
func FetchPersonByEmpID(empID string) (Person, error) {
	row := db.DB.QueryRow(`
		SELECT id, emp_id, name, auth, sex, birth, dpt_id, job, addr, tel, email, state, remark, create_at, update_at
		FROM PERSON WHERE emp_id = ?`, empID)

	var p Person
	err := row.Scan(&p.ID, &p.EmpID, &p.Name, &p.Auth, &p.Sex, &p.Birth,
		&p.DptID, &p.Job, &p.Addr, &p.Tel, &p.Email, &p.State, &p.Remark,
		&p.CreateAt, &p.UpdateAt)
	return p, err
}

// 更新员工部门（按部门名称）
func UpdatePersonDepartment(empID string, deptName string) error {
	_, err := db.DB.Exec(`
		UPDATE person
		SET dpt_id = (SELECT id FROM department WHERE name = ? LIMIT 1),
		    update_at = datetime('now')
		WHERE emp_id = ?`, deptName, empID)
	return err
}

// 更新员工状态（在职/离职）
func UpdatePersonState(empID string, state int) error {
	_, err := db.DB.Exec(`UPDATE person SET state=?, update_at=datetime('now') WHERE emp_id=?`, state, empID)
	return err
}

// 更新员工职位
func UpdatePersonJob(empID string, newJob string) error {
	_, err := db.DB.Exec(`UPDATE person SET job=?, update_at=datetime('now') WHERE emp_id=?`, newJob, empID)
	return err
}
