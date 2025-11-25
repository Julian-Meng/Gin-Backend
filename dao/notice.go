package dao

import (
	"backend/db"
	"backend/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// =============================
// 创建公告
// =============================
func CreateNotice(n *models.Notice) error {
	dbConn := db.GetDB()
	return dbConn.Create(n).Error
}

// =============================
// 更新公告
// =============================
func UpdateNotice(id uint, n models.Notice) error {
	dbConn := db.GetDB()

	updates := map[string]interface{}{
		"title":      n.Title,
		"content":    n.Content,
		"updated_at": time.Now(),
	}

	return dbConn.
		Model(&models.Notice{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// =============================
// 删除公告
// =============================
func DeleteNotice(id uint) error {
	dbConn := db.GetDB()
	return dbConn.Delete(&models.Notice{}, id).Error
}

// =============================
// 分页获取公告
// =============================
func GetAllNotices(page, pageSize int) ([]models.Notice, int64, error) {
	dbConn := db.GetDB()

	offset := (page - 1) * pageSize
	var list []models.Notice
	var total int64

	// count
	dbConn.Model(&models.Notice{}).Count(&total)

	// 查询
	err := dbConn.
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&list).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询公告失败: %v", err)
	}

	return list, total, nil
}

// =============================
// 查询单条公告
// =============================
func GetNoticeByID(id uint) (*models.Notice, error) {
	dbConn := db.GetDB()

	var n models.Notice
	err := dbConn.First(&n, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到公告 ID=%d", id)
		}
		return nil, err
	}

	return &n, nil
}
