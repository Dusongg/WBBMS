package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type BorrowService struct {
	fineService        *FineService
	reservationService *ReservationService
	blacklistService   *BlacklistService
}

// NewBorrowService 创建借还书服务实例
func NewBorrowService() *BorrowService {
	return &BorrowService{
		fineService:        &FineService{},
		reservationService: NewReservationService(),
		blacklistService:   &BlacklistService{},
	}
}

// BorrowBook 借书（增强版）
func (s *BorrowService) BorrowBook(userID, bookID uint, operatorID uint, reservationID *uint) (*model.BorrowRecord, error) {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查读者是否存在且状态正常（根据UserID查询）
	var reader model.Reader
	if err := tx.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		tx.Rollback()
		// 如果读者不存在，自动创建一个
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
			return nil, errors.New("创建读者信息失败")
		}
		// 重新开始事务
		tx = global.GVA_DB.Begin()
	}

	if reader.Status != model.ReaderStatusActive {
		tx.Rollback()
		return nil, errors.New("读者状态异常，无法借书")
	}

	// 2. 检查是否在黑名单
	if reader.IsBlacklisted {
		tx.Rollback()
		return nil, errors.New("您已被列入黑名单，无法借书")
	}

	// 3. 检查是否有未支付罚款
	if reader.UnpaidFine > 0 {
		tx.Rollback()
		return nil, errors.New("您有未支付的罚款，请先支付后再借书")
	}

	// 4. 检查是否有逾期未还图书
	var overdueCount int64
	tx.Model(&model.BorrowRecord{}).
		Where("reader_id = ? AND status = ?", reader.ID, model.BorrowStatusOverdue).
		Count(&overdueCount)

	if overdueCount > 0 {
		blockDays := GlobalConfigService.GetIntConfig(model.ConfigOverdueBlockDays, 7)
		tx.Rollback()
		return nil, fmt.Errorf("您有图书逾期未还，超过%d天将禁止借书", blockDays)
	}

	// 5. 检查图书是否存在
	var book model.Book
	if err := tx.First(&book, bookID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("图书不存在")
	}

	// 6. 检查库存
	if book.AvailableStock <= 0 {
		tx.Rollback()
		return nil, errors.New("图书库存不足")
	}

	// 6.5. 检查是否有其他用户的可用预约（非通过预约借书的情况）
	if reservationID == nil {
		now := time.Now()
		// 查找状态为available且未过期的预约
		var availableReservations []model.Reservation
		err := tx.Where("book_id = ? AND status = ? AND (pickup_deadline IS NULL OR pickup_deadline > ?)",
			bookID, model.ReservationStatusAvailable, now).
			Find(&availableReservations).Error

		if err == nil && len(availableReservations) > 0 {
			// 查找当前用户的预约
			var userReservation *model.Reservation
			for i := range availableReservations {
				if availableReservations[i].ReaderID == reader.ID {
					userReservation = &availableReservations[i]
					break
				}
			}

			if userReservation == nil {
				// 有其他用户的预约，但不是当前用户的
				tx.Rollback()
				return nil, errors.New("该图书已被其他用户预约，请等待预约失效后再试")
			}

			// 如果是当前用户的预约，但没有传递reservationID，自动使用该预约
			// 这样可以简化前端逻辑，用户可以直接点击借阅按钮
			reservationID = &userReservation.ID
		}
	}

	// 7. 检查读者借阅数量
	var borrowCount int64
	tx.Model(&model.BorrowRecord{}).
		Where("reader_id = ? AND status IN (?)", reader.ID, []model.BorrowStatus{
			model.BorrowStatusBorrowed,
			model.BorrowStatusOverdue,
		}).
		Count(&borrowCount)

	maxBorrow := GlobalConfigService.GetIntConfig(model.ConfigMaxBorrowBooks, reader.MaxBorrow)
	if int(borrowCount) >= maxBorrow {
		tx.Rollback()
		return nil, errors.New("已达到最大借阅数量")
	}

	// 8. 检查是否已借阅该书
	var existRecord model.BorrowRecord
	if err := tx.Where("reader_id = ? AND book_id = ? AND status IN (?)", reader.ID, bookID, []model.BorrowStatus{
		model.BorrowStatusBorrowed,
		model.BorrowStatusOverdue,
	}).First(&existRecord).Error; err == nil {
		tx.Rollback()
		return nil, errors.New("您已借阅该书")
	}

	// 9. 如果是通过预约借书，验证预约
	if reservationID != nil {
		var reservation model.Reservation
		if err := tx.First(&reservation, *reservationID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("预约记录不存在")
		}

		if reservation.ReaderID != reader.ID || reservation.BookID != bookID {
			tx.Rollback()
			return nil, errors.New("预约信息不匹配")
		}

		if reservation.Status != model.ReservationStatusAvailable {
			tx.Rollback()
			return nil, errors.New("预约状态异常")
		}

		// 更新预约状态为已完成
		now := time.Now()
		if err := tx.Model(&reservation).Updates(map[string]interface{}{
			"status":         model.ReservationStatusFulfilled,
			"fulfilled_date": now,
		}).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("更新预约状态失败")
		}
	}

	// 10. 创建借阅记录
	now := time.Now()
	borrowDays := GlobalConfigService.GetIntConfig(model.ConfigBorrowDays, reader.BorrowDays)
	maxRenewCount := GlobalConfigService.GetIntConfig(model.ConfigMaxRenewTimes, reader.MaxRenew)

	// 测试模式：借阅期限改为1分钟
	// 生产环境：使用天数
	var dueDate time.Time
	if borrowDays == 0 || borrowDays > 365 {
		// 如果配置为0或异常大值，使用1分钟（测试模式）
		dueDate = now.Add(1 * time.Minute)
	} else {
		dueDate = now.AddDate(0, 0, borrowDays)
	}

	// 判断是否为普通用户借书（userID == operatorID）还是管理员为读者借书
	var borrowStatus model.BorrowStatus
	var shouldUpdateStock bool
	if userID == operatorID {
		// 普通用户自己借书，需要管理员审批
		borrowStatus = model.BorrowStatusPending
		shouldUpdateStock = false // 待审批时不减库存
	} else {
		// 管理员为读者借书，直接借出
		borrowStatus = model.BorrowStatusBorrowed
		shouldUpdateStock = true
	}

	record := model.BorrowRecord{
		ReaderID:      reader.ID,
		BookID:        bookID,
		BorrowDate:    now,
		DueDate:       dueDate,
		Status:        borrowStatus,
		MaxRenewCount: maxRenewCount,
		ReservationID: reservationID,
		OperatorID:    operatorID,
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("创建借阅记录失败", zap.Error(err))
		return nil, errors.New("借书失败")
	}

	// 11. 更新图书库存（仅管理员直接借书时）
	if shouldUpdateStock {
		if err := tx.Model(&book).Update("available_stock", book.AvailableStock-1).Error; err != nil {
			tx.Rollback()
			global.GVA_LOG.Error("更新库存失败", zap.Error(err))
			return nil, errors.New("借书失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("借书失败")
	}

	if borrowStatus == model.BorrowStatusPending {
		global.GVA_LOG.Info("借书申请提交成功", zap.Uint("reader_id", reader.ID), zap.Uint("book_id", bookID))
	} else {
		global.GVA_LOG.Info("借书成功", zap.Uint("reader_id", reader.ID), zap.Uint("book_id", bookID))
	}
	return &record, nil
}

// ApproveBorrowRequest 审批借阅申请
func (s *BorrowService) ApproveBorrowRequest(recordID uint, operatorID uint, approved bool, rejectReason string) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查找借阅记录
	var record model.BorrowRecord
	if err := tx.Preload("Book").Preload("Reader").First(&record, recordID).Error; err != nil {
		tx.Rollback()
		return errors.New("借阅记录不存在")
	}

	// 2. 检查状态是否为待批准
	if record.Status != model.BorrowStatusPending {
		tx.Rollback()
		return errors.New("该借阅申请已处理")
	}

	// 3. 检查图书是否还有库存
	var book model.Book
	if err := tx.First(&book, record.BookID).Error; err != nil {
		tx.Rollback()
		return errors.New("图书不存在")
	}

	if approved {
		// 批准借阅
		if book.AvailableStock <= 0 {
			tx.Rollback()
			return errors.New("图书库存不足，无法批准")
		}

		// 更新借阅记录状态
		if err := tx.Model(&record).Updates(map[string]interface{}{
			"status":      model.BorrowStatusBorrowed,
			"operator_id": operatorID,
		}).Error; err != nil {
			tx.Rollback()
			return errors.New("更新借阅记录失败")
		}

		// 减少图书库存
		if err := tx.Model(&book).Update("available_stock", book.AvailableStock-1).Error; err != nil {
			tx.Rollback()
			return errors.New("更新库存失败")
		}

		global.GVA_LOG.Info("批准借阅申请", zap.Uint("record_id", recordID), zap.Uint("operator_id", operatorID))
	} else {
		// 拒绝借阅
		if err := tx.Model(&record).Updates(map[string]interface{}{
			"status":      model.BorrowStatusRejected,
			"operator_id": operatorID,
			"remark":      rejectReason,
		}).Error; err != nil {
			tx.Rollback()
			return errors.New("更新借阅记录失败")
		}

		global.GVA_LOG.Info("拒绝借阅申请", zap.Uint("record_id", recordID), zap.Uint("operator_id", operatorID))
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("操作失败")
	}

	return nil
}

// CancelBorrowRequest 取消借阅申请（用户自己取消）
func (s *BorrowService) CancelBorrowRequest(recordID uint, userID uint) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查找借阅记录
	var record model.BorrowRecord
	if err := tx.Preload("Reader").First(&record, recordID).Error; err != nil {
		tx.Rollback()
		return errors.New("借阅记录不存在")
	}

	// 2. 验证是否是本人的借阅申请
	if record.Reader.UserID != userID {
		tx.Rollback()
		return errors.New("无权取消此借阅申请")
	}

	// 3. 检查状态是否为待批准
	if record.Status != model.BorrowStatusPending {
		tx.Rollback()
		return errors.New("只能取消待审批的借阅申请")
	}

	// 4. 更新状态为已取消（使用rejected表示取消）
	if err := tx.Model(&record).Updates(map[string]interface{}{
		"status": model.BorrowStatusRejected,
		"remark": "用户主动取消",
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("取消借阅申请失败")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("操作失败")
	}

	global.GVA_LOG.Info("取消借阅申请", zap.Uint("record_id", recordID), zap.Uint("user_id", userID))
	return nil
}

// ReturnBook 还书（增强版）
func (s *BorrowService) ReturnBook(recordID uint, operatorID uint) (*model.BorrowRecord, float64, error) {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查找借阅记录
	var record model.BorrowRecord
	if err := tx.Preload("Book").Preload("Reader").First(&record, recordID).Error; err != nil {
		tx.Rollback()
		return nil, 0, errors.New("借阅记录不存在")
	}

	if record.Status != model.BorrowStatusBorrowed && record.Status != model.BorrowStatusOverdue {
		tx.Rollback()
		return nil, 0, errors.New("该记录已归还")
	}

	// 2. 计算罚款
	fineAmount, overdueDays, err := s.fineService.CalculateOverdueFine(&record)
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 3. 更新借阅记录
	now := time.Now()
	record.ReturnDate = &now
	record.FineAmount = fineAmount
	record.OverdueDays = overdueDays

	if overdueDays > 0 {
		record.Status = model.BorrowStatusReturned // 改为已归还，但有罚款记录
	} else {
		record.Status = model.BorrowStatusReturned
	}

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("更新借阅记录失败", zap.Error(err))
		return nil, 0, errors.New("还书失败")
	}

	// 4. 如果有罚款，创建罚款记录
	if fineAmount > 0 {
		if err := s.fineService.CreateFineRecord(record.ReaderID, record.ID, "overdue", fineAmount, overdueDays, operatorID); err != nil {
			global.GVA_LOG.Error("创建罚款记录失败", zap.Error(err))
			// 不回滚，允许还书成功但罚款记录创建失败
		}
	}

	// 5. 更新图书库存
	if err := tx.Model(&record.Book).Update("available_stock", record.Book.AvailableStock+1).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("更新库存失败", zap.Error(err))
		return nil, 0, errors.New("还书失败")
	}

	// 6. 检查并通知预约者
	go func() {
		if err := s.reservationService.CheckAndNotifyAvailableReservations(record.BookID); err != nil {
			global.GVA_LOG.Error("通知预约者失败", zap.Error(err))
		}
	}()

	if err := tx.Commit().Error; err != nil {
		return nil, 0, errors.New("还书失败")
	}

	global.GVA_LOG.Info("还书成功", zap.Uint("record_id", recordID), zap.Float64("fine", fineAmount))
	return &record, fineAmount, nil
}

// RenewBook 续借（增强版）
func (s *BorrowService) RenewBook(recordID, readerID uint) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查找借阅记录
	var record model.BorrowRecord
	if err := tx.Preload("Reader").First(&record, recordID).Error; err != nil {
		tx.Rollback()
		return errors.New("借阅记录不存在")
	}

	// 2. 验证是否是本人的借阅
	if record.ReaderID != readerID {
		tx.Rollback()
		return errors.New("无权续借此记录")
	}

	// 3. 检查状态
	if record.Status != model.BorrowStatusBorrowed {
		tx.Rollback()
		return errors.New("只能续借未归还的图书")
	}

	// 4. 检查是否已逾期
	if time.Now().After(record.DueDate) {
		tx.Rollback()
		return errors.New("图书已逾期，无法续借，请先归还")
	}

	// 5. 检查续借次数
	if record.RenewCount >= record.MaxRenewCount {
		tx.Rollback()
		return errors.New("已达到最大续借次数")
	}

	// 6. 检查该书是否有人预约
	var reservationCount int64
	tx.Model(&model.Reservation{}).
		Where("book_id = ? AND status IN (?)", record.BookID, []model.ReservationStatus{
			model.ReservationStatusPending,
			model.ReservationStatusAvailable,
		}).
		Count(&reservationCount)

	if reservationCount > 0 {
		tx.Rollback()
		return errors.New("该图书已有人预约，无法续借")
	}

	// 7. 更新借阅记录
	renewDays := GlobalConfigService.GetIntConfig(model.ConfigRenewDays, record.Reader.RenewDays)
	record.DueDate = record.DueDate.AddDate(0, 0, renewDays)
	record.RenewCount++

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("更新借阅记录失败", zap.Error(err))
		return errors.New("续借失败")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("续借失败")
	}

	global.GVA_LOG.Info("续借成功", zap.Uint("record_id", recordID), zap.Time("new_due_date", record.DueDate))
	return nil
}

// CheckOverdueRecords 检查并更新逾期记录
// 定时任务调用
func (s *BorrowService) CheckOverdueRecords() error {
	// 查找所有未归还且已过期的借阅记录
	now := time.Now()
	var overdueRecords []model.BorrowRecord

	if err := global.GVA_DB.Where("status = ? AND due_date < ?", model.BorrowStatusBorrowed, now).
		Find(&overdueRecords).Error; err != nil {
		return err
	}

	if len(overdueRecords) == 0 {
		return nil
	}

	// 批量更新状态为逾期
	recordIDs := make([]uint, len(overdueRecords))
	for i, r := range overdueRecords {
		recordIDs[i] = r.ID
	}

	if err := global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("id IN ?", recordIDs).
		Update("status", model.BorrowStatusOverdue).Error; err != nil {
		return err
	}

	global.GVA_LOG.Info("更新逾期记录", zap.Int("count", len(overdueRecords)))

	// 发送逾期提醒通知
	messageService := &MessageService{}
	for _, record := range overdueRecords {
		// 预加载读者和图书信息
		if err := global.GVA_DB.Preload("Reader").Preload("Book").First(&record, record.ID).Error; err != nil {
			global.GVA_LOG.Error("预加载借阅记录失败", zap.Error(err))
			continue
		}

		// 计算逾期时间
		overdueDuration := time.Since(record.DueDate)
		overdueDays := int(overdueDuration.Hours() / 24)
		overdueMinutes := int(overdueDuration.Minutes())

		// 如果逾期时间小于1天，显示分钟数（测试模式）
		var content string
		if overdueDays < 1 {
			if overdueMinutes < 1 {
				overdueSeconds := int(overdueDuration.Seconds())
				content = fmt.Sprintf("您借阅的《%s》已逾期 %d 秒，请尽快归还。逾期将产生罚款。",
					record.Book.Title, overdueSeconds)
			} else {
				content = fmt.Sprintf("您借阅的《%s》已逾期 %d 分钟，请尽快归还。逾期将产生罚款。",
					record.Book.Title, overdueMinutes)
			}
		} else {
			content = fmt.Sprintf("您借阅的《%s》已逾期 %d 天，请尽快归还。逾期将产生罚款。",
				record.Book.Title, overdueDays)
		}

		// 发送站内消息
		title := "⚠️ 图书已逾期"

		relatedID := record.ID
		if err := messageService.CreateMessage(
			record.Reader.UserID,
			model.MessageTypeOverdue,
			title,
			content,
			&relatedID,
			"borrow",
		); err != nil {
			global.GVA_LOG.Error("发送逾期提醒消息失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("发送逾期提醒消息成功",
				zap.Uint("reader_id", record.ReaderID),
				zap.Uint("book_id", record.BookID),
				zap.Int("overdue_days", overdueDays))
		}
	}

	return nil
}

// SendDueReminders 发送到期提醒
// 定时任务调用
func (s *BorrowService) SendDueReminders() error {
	// 获取提醒天数配置
	reminderDays := GlobalConfigService.GetIntConfig(model.ConfigOverdueReminderDays, 3)

	// 查找即将到期的借阅记录
	// 测试模式：提前30秒提醒
	// 生产环境：使用天数
	var startDate, endDate time.Time
	now := time.Now()
	if reminderDays == 0 || reminderDays > 365 {
		// 如果配置为0或异常大值，使用30秒（测试模式）
		startDate = now.Add(30 * time.Second)
		endDate = startDate.Add(1 * time.Minute)
	} else {
		startDate = now.AddDate(0, 0, reminderDays)
		endDate = startDate.AddDate(0, 0, 1)
	}

	var dueRecords []model.BorrowRecord
	if err := global.GVA_DB.Where("status = ? AND due_date >= ? AND due_date < ?",
		model.BorrowStatusBorrowed, startDate, endDate).
		Preload("Reader").
		Preload("Book").
		Find(&dueRecords).Error; err != nil {
		return err
	}

	for _, record := range dueRecords {
		// TODO: 发送提醒通知（邮件、短信、站内信等）
		global.GVA_LOG.Info("发送到期提醒",
			zap.Uint("reader_id", record.ReaderID),
			zap.Uint("book_id", record.BookID),
			zap.Time("due_date", record.DueDate))
	}

	return nil
}
