package models

import (
	"backend/db"
	"database/sql"
	"fmt"
	"log"
)

// =============================
// Notice 数据结构
// =============================
type Notice struct {
	ID        int            `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Publisher string         `json:"publisher"`
	CreateAt  string         `json:"create_at"`
	UpdateAt  sql.NullString `json:"update_at"`
}

// =============================
// 创建公告
// =============================
func CreateNotice(n Notice) error {
	_, err := db.DB.Exec(`
		INSERT INTO NOTICE (title, content, publisher, create_at)
		VALUES (?, ?, ?, datetime('now'))
	`, n.Title, n.Content, n.Publisher)
	return err
}

// =============================
// 更新公告
// =============================
func UpdateNotice(id string, n Notice) error {
	_, err := db.DB.Exec(`
		UPDATE NOTICE
		SET title=?, content=?, update_at=datetime('now')
		WHERE id=?
	`, n.Title, n.Content, id)
	return err
}

// =============================
// 删除公告
// =============================
func DeleteNotice(id string) error {
	_, err := db.DB.Exec(`DELETE FROM NOTICE WHERE id=?`, id)
	return err
}

// =============================
// 获取所有公告（分页）
// =============================
func GetAllNotices(page, pageSize int) ([]Notice, int, error) {
	offset := (page - 1) * pageSize

	rows, err := db.DB.Query(`
		SELECT id, title, content, publisher, create_at, update_at
		FROM NOTICE
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询公告失败: %v", err)
	}
	defer rows.Close()

	var notices []Notice
	for rows.Next() {
		var n Notice
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Publisher, &n.CreateAt, &n.UpdateAt); err != nil {
			log.Println("⚠️ 扫描公告出错:", err)
			continue
		}
		notices = append(notices, n)
	}

	var total int
	_ = db.DB.QueryRow(`SELECT COUNT(*) FROM NOTICE`).Scan(&total)

	return notices, total, nil
}

// =============================
// 获取单条公告
// =============================
func GetNoticeByID(id string) (Notice, error) {
	row := db.DB.QueryRow(`
		SELECT id, title, content, publisher, create_at, update_at
		FROM NOTICE WHERE id=?
	`, id)

	var n Notice
	err := row.Scan(&n.ID, &n.Title, &n.Content, &n.Publisher, &n.CreateAt, &n.UpdateAt)
	return n, err
}
