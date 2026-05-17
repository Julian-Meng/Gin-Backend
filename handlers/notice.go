package handlers

import (
	"backend/dao"
	"backend/middlewares/errorx"
	"backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAllNotices godoc
// @Summary 获取公告列表
// @Tags notice
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} NoticeListResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/notice [get]
func GetAllNotices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	list, total, err := dao.GetAllNotices(page, pageSize)
	if err != nil {
		errorx.Internal(c, "获取公告列表失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "ok",
		"data":  list,
		"total": total,
	})
}

// GetNoticeByID godoc
// @Summary 获取公告详情
// @Tags notice
// @Produce json
// @Security BearerAuth
// @Param id path int true "公告ID"
// @Success 200 {object} NoticeDetailResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /api/admin/notice/{id} [get]
func GetNoticeByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的公告 ID", err)
		return
	}

	n, err := dao.GetNoticeByID(uint(id64))
	if err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "公告不存在", err)
			return
		}
		errorx.Internal(c, "查询公告失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": n,
	})
}

// CreateNotice godoc
// @Summary 创建公告
// @Tags notice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Notice true "公告信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/notice [post]
func CreateNotice(c *gin.Context) {
	var req models.Notice

	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Publisher = strings.TrimSpace(req.Publisher)

	if req.Title == "" {
		errorx.BadRequest(c, "公告标题不能为空", nil)
		return
	}

	if err := dao.CreateNotice(&req); err != nil {
		errorx.Internal(c, "创建公告失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "公告创建成功",
	})
}

// UpdateNotice godoc
// @Summary 更新公告
// @Tags notice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "公告ID"
// @Param request body models.Notice true "公告信息"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/notice/{id} [put]
func UpdateNotice(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的公告 ID", err)
		return
	}

	var req models.Notice
	if err := c.ShouldBindJSON(&req); err != nil {
		errorx.BadRequest(c, "参数解析失败", err)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Publisher = strings.TrimSpace(req.Publisher)

	if req.Title == "" {
		errorx.BadRequest(c, "公告标题不能为空", nil)
		return
	}

	if err := dao.UpdateNotice(uint(id64), req); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "公告不存在", err)
			return
		}
		errorx.Internal(c, "公告更新失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告更新成功"})
}

// DeleteNotice godoc
// @Summary 删除公告
// @Tags notice
// @Produce json
// @Security BearerAuth
// @Param id path int true "公告ID"
// @Success 200 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/admin/notice/{id} [delete]
func DeleteNotice(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		errorx.BadRequest(c, "无效的公告 ID", err)
		return
	}

	if err := dao.DeleteNotice(uint(id64)); err != nil {
		if dao.IsNotFound(err) {
			errorx.NotFound(c, "公告不存在", err)
			return
		}
		errorx.Internal(c, "删除公告失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "公告删除成功"})
}
