package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReservationApi struct{}

var reservationService = &service.ReservationService{}

// CreateReservation 创建预约
func (r *ReservationApi) CreateReservation(c *gin.Context) {
	var req struct {
		BookID uint `json:"book_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 查找或创建读者记录
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		// 如果读者不存在，自动创建
		reader = model.Reader{
			UserID:          userID,
			ReaderNo:        fmt.Sprintf("R%d", userID),
			Status:          model.ReaderStatusActive,
			MaxBorrow:       5,
			BorrowDays:      30,
			MaxRenew:        2,
			RenewDays:       15,
			MaxReservations: 3,
			TotalFine:       0,
			UnpaidFine:      0,
			IsBlacklisted:   false,
		}
		if err := global.GVA_DB.Create(&reader).Error; err != nil {
			c.JSON(200, response.FailWithMessage("创建读者信息失败"))
			return
		}
	}

	// 创建预约
	reservation, err := reservationService.CreateReservation(reader.ID, req.BookID)
	if err != nil {
		global.GVA_LOG.Error("创建预约失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"message":        "预约成功",
		"queue_position": reservation.QueuePosition,
		"reservation_id": reservation.ID,
	}))
}

// CancelReservation 取消预约
func (r *ReservationApi) CancelReservation(c *gin.Context) {
	reservationIDStr := c.Param("id")
	reservationID, err := strconv.ParseUint(reservationIDStr, 10, 32)
	if err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 查找读者ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 取消预约
	if err := reservationService.CancelReservation(uint(reservationID), reader.ID); err != nil {
		global.GVA_LOG.Error("取消预约失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("取消预约成功"))
}

// GetMyReservations 获取我的预约列表
func (r *ReservationApi) GetMyReservations(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 查找读者ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 获取预约列表
	reservations, err := reservationService.GetReaderReservations(reader.ID)
	if err != nil {
		global.GVA_LOG.Error("获取预约列表失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取预约列表失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"list": reservations,
	}))
}

// GetReservationList 获取预约列表（管理员）
func (r *ReservationApi) GetReservationList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var reservations []model.Reservation
	var total int64
	db := global.GVA_DB.Model(&model.Reservation{}).
		Preload("Reader").
		Preload("Reader.User").
		Preload("Book")

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + pageInfo.Keyword + "%"
		db = db.Joins("JOIN books ON reservations.book_id = books.id").
			Joins("JOIN readers ON reservations.reader_id = readers.id").
			Where("books.title LIKE ? OR readers.reader_no LIKE ?", keyword, keyword)
	}

	// 状态筛选
	status := c.Query("status")
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Order("reservations.created_at DESC").
		Limit(pageInfo.PageSize).
		Offset(offset).
		Find(&reservations).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     reservations,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}
