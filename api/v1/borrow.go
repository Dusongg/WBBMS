package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BorrowApi struct{}

var borrowService = service.NewBorrowService()

// BorrowBook 借书（增强版）
func (b *BorrowApi) BorrowBook(c *gin.Context) {
	var req struct {
		ReaderID      *uint `json:"reader_id"` // 可选，管理员可以指定读者ID
		BookID        uint  `json:"book_id" binding:"required"`
		ReservationID *uint `json:"reservation_id"` // 可选，如果是通过预约借书
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

	// 确定用户ID：如果req.ReaderID为空，使用当前用户ID
	// 注意：这里传递的是UserID，服务层会根据UserID查找或创建Reader记录
	var targetUserID uint
	if req.ReaderID != nil && *req.ReaderID > 0 {
		// 管理员为其他读者借书（这里的ReaderID实际上是UserID）
		targetUserID = *req.ReaderID
	} else {
		// 普通用户为自己借书
		targetUserID = userID
	}

	// 调用增强的借书服务（传递UserID）
	record, err := borrowService.BorrowBook(targetUserID, req.BookID, userID, req.ReservationID)
	if err != nil {
		global.GVA_LOG.Error("借书失败", zap.Error(err))

		// 检查是否是库存不足错误，返回特殊code以便前端处理
		if err.Error() == "图书库存不足" {
			c.JSON(200, response.Response{
				Code: 4001, // 特殊code：库存不足，建议预约
				Msg:  "图书库存不足",
				Data: gin.H{
					"book_id": req.BookID,
					"suggest": "reserve", // 建议预约
				},
			})
			return
		}

		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"message": "借书成功",
		"id":      record.ID,
	}))
}

// ApproveBorrowRequest 审批借阅申请（管理员）
func (b *BorrowApi) ApproveBorrowRequest(c *gin.Context) {
	var req struct {
		RecordID     uint   `json:"record_id" binding:"required"`
		Approved     bool   `json:"approved"`
		RejectReason string `json:"reject_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("绑定参数失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(fmt.Sprintf("参数错误: %v", err)))
		return
	}

	global.GVA_LOG.Info("审批请求参数",
		zap.Uint("record_id", req.RecordID),
		zap.Bool("approved", req.Approved),
		zap.String("reject_reason", req.RejectReason))

	// 获取当前用户ID（管理员）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 调用审批服务
	if err := borrowService.ApproveBorrowRequest(req.RecordID, userID, req.Approved, req.RejectReason); err != nil {
		global.GVA_LOG.Error("审批借阅申请失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	if req.Approved {
		c.JSON(200, response.OkWithMessage("已批准借阅申请"))
	} else {
		c.JSON(200, response.OkWithMessage("已拒绝借阅申请"))
	}
}

// CancelBorrowRequest 取消借阅申请（用户自己取消）
func (b *BorrowApi) CancelBorrowRequest(c *gin.Context) {
	var req struct {
		RecordID uint `json:"record_id" binding:"required"`
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

	// 调用取消服务
	if err := borrowService.CancelBorrowRequest(req.RecordID, userID); err != nil {
		global.GVA_LOG.Error("取消借阅申请失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("已取消借阅申请"))
}

// ReturnBook 还书（增强版）
func (b *BorrowApi) ReturnBook(c *gin.Context) {
	var req struct {
		ID       uint `json:"id"`        // 前端传递的参数
		RecordID uint `json:"record_id"` // 兼容旧参数
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 确定借阅记录ID（优先使用ID，其次使用RecordID）
	recordID := req.ID
	if recordID == 0 {
		recordID = req.RecordID
	}
	if recordID == 0 {
		c.JSON(200, response.FailWithMessage("借阅记录ID不能为空"))
		return
	}

	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 调用增强的还书服务
	record, fineAmount, err := borrowService.ReturnBook(recordID, userID)
	if err != nil {
		global.GVA_LOG.Error("还书失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	result := gin.H{
		"message": "还书成功",
	}

	if fineAmount > 0 {
		result["message"] = "还书成功，产生逾期费用"
		result["fine_amount"] = fineAmount
		result["overdue_days"] = record.OverdueDays
	}

	c.JSON(200, response.OkWithData(result))
}

// RenewBook 续借（增强版）
func (b *BorrowApi) RenewBook(c *gin.Context) {
	var req struct {
		RecordID uint `json:"record_id" binding:"required"`
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

	// 查找读者ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 调用增强的续借服务
	if err := borrowService.RenewBook(req.RecordID, reader.ID); err != nil {
		global.GVA_LOG.Error("续借失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("续借成功"))
}

// GetBorrowList 获取借阅记录列表（管理员）
func (b *BorrowApi) GetBorrowList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var records []model.BorrowRecord
	var total int64
	db := global.GVA_DB.Model(&model.BorrowRecord{}).Preload("Reader.User").Preload("Book")

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + pageInfo.Keyword + "%"
		db = db.Joins("JOIN books ON borrow_records.book_id = books.id").
			Where("books.title LIKE ? OR books.isbn LIKE ?", keyword, keyword)
	}

	// 状态筛选
	status := c.Query("status")
	if status != "" {
		db = db.Where("borrow_records.status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Order("borrow_records.created_at DESC").Limit(pageInfo.PageSize).Offset(offset).Find(&records).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     records,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// GetMyBorrowList 获取我的借阅记录（增强版）
func (b *BorrowApi) GetMyBorrowList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}

	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	// 查找读者的ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		// 如果读者不存在，返回空列表（新用户还没有借阅记录）
		c.JSON(200, response.OkWithDetailed(response.PageResult{
			List:     []model.BorrowRecord{},
			Total:    0,
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
		}, "获取成功"))
		return
	}

	var records []model.BorrowRecord
	var total int64
	db := global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("reader_id = ?", reader.ID).
		Preload("Book").
		Preload("Book.Categories")

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
	if err := db.Order("created_at DESC").Limit(pageInfo.PageSize).Offset(offset).Find(&records).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	// 添加读者信息到响应
	c.JSON(200, response.OkWithDetailed(gin.H{
		"list":           records,
		"total":          total,
		"page":           pageInfo.Page,
		"pageSize":       pageInfo.PageSize,
		"unpaid_fine":    reader.UnpaidFine,
		"is_blacklisted": reader.IsBlacklisted,
	}, "获取成功"))
}

// GetBorrowStatistics 获取借阅统计信息（我的）
func (b *BorrowApi) GetBorrowStatistics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}

	// 查找读者的ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		// 如果读者不存在，返回默认统计数据（新用户）
		c.JSON(200, response.OkWithData(gin.H{
			"borrowing_count":    0,
			"overdue_count":      0,
			"total_borrow_count": 0,
			"reservation_count":  0,
			"max_borrow":         5,
			"unpaid_fine":        0,
			"total_fine":         0,
			"is_blacklisted":     false,
		}))
		return
	}

	// 统计借阅中的图书
	var borrowingCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("reader_id = ? AND status IN (?)", reader.ID, []model.BorrowStatus{
			model.BorrowStatusBorrowed,
			model.BorrowStatusOverdue,
		}).
		Count(&borrowingCount)

	// 统计逾期的图书
	var overdueCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("reader_id = ? AND status = ?", reader.ID, model.BorrowStatusOverdue).
		Count(&overdueCount)

	// 统计历史借阅总数
	var totalBorrowCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("reader_id = ?", reader.ID).
		Count(&totalBorrowCount)

	// 统计预约数量
	var reservationCount int64
	global.GVA_DB.Model(&model.Reservation{}).
		Where("reader_id = ? AND status IN (?)", reader.ID, []model.ReservationStatus{
			model.ReservationStatusPending,
			model.ReservationStatusAvailable,
		}).
		Count(&reservationCount)

	c.JSON(200, response.OkWithData(gin.H{
		"borrowing_count":    borrowingCount,
		"overdue_count":      overdueCount,
		"total_borrow_count": totalBorrowCount,
		"reservation_count":  reservationCount,
		"max_borrow":         reader.MaxBorrow,
		"unpaid_fine":        reader.UnpaidFine,
		"total_fine":         reader.TotalFine,
		"is_blacklisted":     reader.IsBlacklisted,
	}))
}

// PayFine 支付罚款（根据借阅记录ID）
func (b *BorrowApi) PayFine(c *gin.Context) {
	var req struct {
		RecordID uint `json:"record_id" binding:"required"`
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

	// 查找读者的ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 查找借阅记录
	var borrowRecord model.BorrowRecord
	if err := global.GVA_DB.Where("id = ? AND reader_id = ?", req.RecordID, reader.ID).First(&borrowRecord).Error; err != nil {
		c.JSON(200, response.FailWithMessage("借阅记录不存在"))
		return
	}

	// 查找该借阅记录的罚款记录
	var fineRecord model.FineRecord
	if err := global.GVA_DB.Where("borrow_record_id = ?", req.RecordID).First(&fineRecord).Error; err != nil {
		// 如果没有罚款记录，先计算罚款金额
		fineService := &service.FineService{}
		fineAmount, overdueDays, err := fineService.CalculateOverdueFine(&borrowRecord)
		if err != nil {
			global.GVA_LOG.Error("计算罚款失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("计算罚款失败"))
			return
		}

		if fineAmount <= 0 {
			c.JSON(200, response.FailWithMessage("无需支付罚款"))
			return
		}

		// 创建罚款记录并直接标记为已支付（因为用户提前支付）
		paidDate := time.Now()
		fineRecord = model.FineRecord{
			ReaderID:       reader.ID,
			BorrowRecordID: req.RecordID,
			FineType:       "overdue",
			Amount:         fineAmount,
			PaidAmount:     fineAmount,           // 已支付全部金额
			Status:         model.FineStatusPaid, // 直接标记为已支付
			OverdueDays:    overdueDays,
			FineDate:       time.Now(),
			PaidDate:       &paidDate,
			OperatorID:     userID,
			Remark:         "用户提前支付",
		}

		if err := global.GVA_DB.Create(&fineRecord).Error; err != nil {
			global.GVA_LOG.Error("创建罚款记录失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("创建罚款记录失败"))
			return
		}

		// 更新读者的罚款金额（已支付，所以只更新总罚款，不更新未支付罚款）
		if err := global.GVA_DB.Model(&model.Reader{}).Where("id = ?", reader.ID).
			UpdateColumn("total_fine", gorm.Expr("total_fine + ?", fineAmount)).
			Error; err != nil {
			global.GVA_LOG.Error("更新读者罚款金额失败", zap.Error(err))
		}

		c.JSON(200, response.OkWithData(gin.H{
			"message":     "支付成功",
			"fine_amount": fineAmount,
		}))
		return
	}

	// 如果罚款记录已存在，检查状态
	if fineRecord.Status == model.FineStatusPaid {
		c.JSON(200, response.FailWithMessage("罚款已支付"))
		return
	}

	if fineRecord.Status == model.FineStatusWaived {
		c.JSON(200, response.FailWithMessage("罚款已豁免"))
		return
	}

	// 支付未支付的罚款
	fineService := &service.FineService{}
	unpaidAmount := fineRecord.Amount - fineRecord.PaidAmount
	if err := fineService.PayFine(fineRecord.ID, unpaidAmount, userID); err != nil {
		global.GVA_LOG.Error("支付罚款失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"message":     "支付成功",
		"fine_amount": unpaidAmount,
	}))
}

// GetFineByRecord 根据借阅记录ID获取罚款信息
func (b *BorrowApi) GetFineByRecord(c *gin.Context) {
	recordIDStr := c.Query("record_id")
	if recordIDStr == "" {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var recordID uint
	if _, err := fmt.Sscanf(recordIDStr, "%d", &recordID); err != nil {
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

	// 查找读者的ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 查找借阅记录
	var borrowRecord model.BorrowRecord
	if err := global.GVA_DB.Where("id = ? AND reader_id = ?", recordID, reader.ID).First(&borrowRecord).Error; err != nil {
		c.JSON(200, response.FailWithMessage("借阅记录不存在"))
		return
	}

	// 查找罚款记录
	var fineRecord model.FineRecord
	if err := global.GVA_DB.Where("borrow_record_id = ?", recordID).First(&fineRecord).Error; err != nil {
		// 如果没有罚款记录，计算罚款金额
		fineService := &service.FineService{}
		fineAmount, overdueDays, err := fineService.CalculateOverdueFine(&borrowRecord)
		if err != nil {
			global.GVA_LOG.Error("计算罚款失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("计算罚款失败"))
			return
		}

		c.JSON(200, response.OkWithData(gin.H{
			"fine_amount":  fineAmount,
			"overdue_days": overdueDays,
			"paid_amount":  0,
			"status":       "unpaid",
			"need_pay":     fineAmount > 0,
		}))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"fine_id":      fineRecord.ID,
		"fine_amount":  fineRecord.Amount,
		"paid_amount":  fineRecord.PaidAmount,
		"status":       fineRecord.Status,
		"overdue_days": fineRecord.OverdueDays,
		"need_pay":     fineRecord.Status == model.FineStatusUnpaid && (fineRecord.Amount-fineRecord.PaidAmount) > 0,
	}))
}
