package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ReservationService struct {
	messageService *MessageService
}

// NewReservationService 创建预约服务实例
func NewReservationService() *ReservationService {
	return &ReservationService{
		messageService: &MessageService{},
	}
}

// CreateReservation 创建预约
func (s *ReservationService) CreateReservation(readerID, bookID uint) (*model.Reservation, error) {
	// 开始事务
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查读者是否存在且状态正常
	var reader model.Reader
	if err := tx.First(&reader, readerID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("读者不存在")
	}

	if reader.Status != model.ReaderStatusActive {
		tx.Rollback()
		return nil, errors.New("读者状态异常，无法预约")
	}

	// 2. 检查是否在黑名单
	if reader.IsBlacklisted {
		tx.Rollback()
		return nil, errors.New("您已被列入黑名单，无法预约图书")
	}

	// 3. 检查预约数量限制
	var reservationCount int64
	tx.Model(&model.Reservation{}).
		Where("reader_id = ? AND status IN (?)", readerID, []model.ReservationStatus{
			model.ReservationStatusPending,
			model.ReservationStatusAvailable,
		}).
		Count(&reservationCount)

	maxReservations := GlobalConfigService.GetIntConfig(model.ConfigMaxReservations, reader.MaxReservations)
	if int(reservationCount) >= maxReservations {
		tx.Rollback()
		return nil, errors.New("已达到最大预约数量")
	}

	// 4. 检查是否已预约该书
	var existReservation model.Reservation
	if err := tx.Where("reader_id = ? AND book_id = ? AND status IN (?)", readerID, bookID, []model.ReservationStatus{
		model.ReservationStatusPending,
		model.ReservationStatusAvailable,
	}).First(&existReservation).Error; err == nil {
		tx.Rollback()
		return nil, errors.New("您已预约该书")
	}

	// 5. 检查是否已借阅该书
	var borrowRecord model.BorrowRecord
	if err := tx.Where("reader_id = ? AND book_id = ? AND status IN (?)", readerID, bookID, []model.BorrowStatus{
		model.BorrowStatusBorrowed,
		model.BorrowStatusOverdue,
	}).First(&borrowRecord).Error; err == nil {
		tx.Rollback()
		return nil, errors.New("您已借阅该书，无需预约")
	}

	// 6. 检查图书是否存在
	var book model.Book
	if err := tx.First(&book, bookID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("图书不存在")
	}

	// 7. 计算队列位置
	var queuePosition int64
	tx.Model(&model.Reservation{}).
		Where("book_id = ? AND status = ?", bookID, model.ReservationStatusPending).
		Count(&queuePosition)

	// 8. 创建预约记录
	now := time.Now()
	reservation := model.Reservation{
		ReaderID:      readerID,
		BookID:        bookID,
		Status:        model.ReservationStatusPending,
		ReserveDate:   now,
		QueuePosition: int(queuePosition) + 1,
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("创建预约失败", zap.Error(err))
		return nil, errors.New("预约失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("预约失败")
	}

	global.GVA_LOG.Info("预约成功", zap.Uint("reader_id", readerID), zap.Uint("book_id", bookID))
	return &reservation, nil
}

// CancelReservation 取消预约
func (s *ReservationService) CancelReservation(reservationID, readerID uint) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找预约记录
	var reservation model.Reservation
	if err := tx.First(&reservation, reservationID).Error; err != nil {
		tx.Rollback()
		return errors.New("预约记录不存在")
	}

	// 验证是否是本人的预约
	if reservation.ReaderID != readerID {
		tx.Rollback()
		return errors.New("无权取消此预约")
	}

	// 只能取消pending或available状态的预约
	if reservation.Status != model.ReservationStatusPending && reservation.Status != model.ReservationStatusAvailable {
		tx.Rollback()
		return errors.New("该预约无法取消")
	}

	// 更新预约状态
	now := time.Now()
	if err := tx.Model(&reservation).Updates(map[string]interface{}{
		"status":         model.ReservationStatusCancelled,
		"fulfilled_date": now,
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("取消预约失败")
	}

	// 更新后续预约的队列位置
	if reservation.Status == model.ReservationStatusPending {
		tx.Model(&model.Reservation{}).
			Where("book_id = ? AND status = ? AND queue_position > ?",
				reservation.BookID, model.ReservationStatusPending, reservation.QueuePosition).
			UpdateColumn("queue_position", gorm.Expr("queue_position - 1"))
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("取消预约失败")
	}

	global.GVA_LOG.Info("取消预约成功", zap.Uint("reservation_id", reservationID))
	return nil
}

// CheckAndNotifyAvailableReservations 检查并通知可取书的预约
// 当图书归还时调用
func (s *ReservationService) CheckAndNotifyAvailableReservations(bookID uint) error {
	// 查找该书的第一个pending预约，同时加载关联的Reader和Book信息
	var reservation model.Reservation
	if err := global.GVA_DB.Preload("Reader").Preload("Book").
		Where("book_id = ? AND status = ?", bookID, model.ReservationStatusPending).
		Order("queue_position ASC").
		First(&reservation).Error; err != nil {
		// 没有预约，正常情况
		return nil
	}

	// 检查图书是否有可用库存
	var book model.Book
	if err := global.GVA_DB.First(&book, bookID).Error; err != nil {
		return err
	}

	if book.AvailableStock <= 0 {
		// 没有可用库存，无法通知
		return nil
	}

	// 更新预约状态为可取书
	now := time.Now()
	pickupDays := GlobalConfigService.GetIntConfig(model.ConfigReservationPickupDays, 3)
	pickupDeadline := now.AddDate(0, 0, pickupDays)

	if err := global.GVA_DB.Model(&reservation).Updates(map[string]interface{}{
		"status":          model.ReservationStatusAvailable,
		"available_date":  now,
		"pickup_deadline": pickupDeadline,
	}).Error; err != nil {
		global.GVA_LOG.Error("更新预约状态失败", zap.Error(err))
		return err
	}

	// 发送站内消息给读者
	if reservation.Reader.UserID > 0 {
		bookTitle := reservation.Book.Title
		if err := s.messageService.SendReservationAvailableMessage(
			reservation.Reader.UserID,
			bookTitle,
			reservation.ID,
			pickupDays,
		); err != nil {
			global.GVA_LOG.Error("发送预约通知消息失败", zap.Error(err))
			// 不影响主流程，继续执行
		}
	}

	global.GVA_LOG.Info("预约图书已可取", zap.Uint("reservation_id", reservation.ID), zap.Uint("reader_id", reservation.ReaderID))

	return nil
}

// CheckExpiredReservations 检查并处理过期的预约
// 定时任务调用
func (s *ReservationService) CheckExpiredReservations() error {
	// 查找所有过期的available预约
	now := time.Now()
	var expiredReservations []model.Reservation
	if err := global.GVA_DB.Where("status = ? AND pickup_deadline < ?",
		model.ReservationStatusAvailable, now).
		Find(&expiredReservations).Error; err != nil {
		return err
	}

	if len(expiredReservations) == 0 {
		return nil
	}

	// 批量更新为expired
	reservationIDs := make([]uint, len(expiredReservations))
	for i, r := range expiredReservations {
		reservationIDs[i] = r.ID
	}

	if err := global.GVA_DB.Model(&model.Reservation{}).
		Where("id IN ?", reservationIDs).
		Updates(map[string]interface{}{
			"status":         model.ReservationStatusExpired,
			"fulfilled_date": now,
		}).Error; err != nil {
		return err
	}

	global.GVA_LOG.Info("处理过期预约", zap.Int("count", len(expiredReservations)))

	// 对于每本书，通知下一个预约者
	bookIDs := make(map[uint]bool)
	for _, r := range expiredReservations {
		bookIDs[r.BookID] = true
	}

	for bookID := range bookIDs {
		s.CheckAndNotifyAvailableReservations(bookID)
	}

	return nil
}

// GetReaderReservations 获取读者的预约列表
func (s *ReservationService) GetReaderReservations(readerID uint, status ...model.ReservationStatus) ([]model.Reservation, error) {
	var reservations []model.Reservation
	query := global.GVA_DB.Where("reader_id = ?", readerID).
		Preload("Book").
		Order("created_at DESC")

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}

	if err := query.Find(&reservations).Error; err != nil {
		return nil, err
	}

	return reservations, nil
}
