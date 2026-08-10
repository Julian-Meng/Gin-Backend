package dao

import (
	"backend/internal/db"
	"backend/internal/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CreateNotice 创建公告
func CreateNotice(n *model.Notice) error {
	dbConn := db.GetDB()
	return dbConn.Create(n).Error
}

// UpdateNotice 更新公告
func UpdateNotice(id uint, n model.Notice) error {
	dbConn := db.GetDB()

	updates := map[string]interface{}{
		"title":      n.Title,
		"content":    n.Content,
		"updated_at": time.Now(),
	}

	res := dbConn.
		Model(&model.Notice{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("更新公告失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到公告 ID=%d", id))
	}

	return nil
}

// DeleteNotice 删除公告
func DeleteNotice(id uint) error {
	dbConn := db.GetDB()
	res := dbConn.Delete(&model.Notice{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除公告失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return NotFound(fmt.Sprintf("未找到公告 ID=%d", id))
	}

	return nil
}

// GetAllNotices 分页获取公告
func GetAllNotices(page, pageSize int) ([]model.Notice, int64, error) {
	dbConn := db.GetDB()

	offset := (page - 1) * pageSize
	var list []model.Notice
	var total int64

	// count
	if err := dbConn.Model(&model.Notice{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计公告总数失败: %w", err)
	}

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

// GetNoticeByID 查询单条公告
func GetNoticeByID(id uint) (*model.Notice, error) {
	dbConn := db.GetDB()

	var n model.Notice
	err := dbConn.First(&n, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound(fmt.Sprintf("未找到公告 ID=%d", id))
		}
		return nil, fmt.Errorf("查询公告失败: %w", err)
	}

	return &n, nil
}
