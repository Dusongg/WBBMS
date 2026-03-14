package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type BlacklistService struct{}

// AddToBlacklist 添加到黑名单
func (s *BlacklistService) AddToBlacklist(ctx context.Context, readerID uint, reason model.BlacklistReason, description string, endDate *time.Time, operatorID uint) error {
	tx := global.DB(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已在黑名单
	var existBlacklist model.Blacklist
	if err := tx.Where("reader_id = ? AND status = ?", readerID, model.BlacklistStatusActive).
		First(&existBlacklist).Error; err == nil {
		tx.Rollback()
		return errors.New("该读者已在黑名单中")
	}

	// 创建黑名单记录
	blacklist := model.Blacklist{
		ReaderID:    readerID,
		Reason:      reason,
		Description: description,
		Status:      model.BlacklistStatusActive,
		StartDate:   time.Now(),
		EndDate:     endDate,
		OperatorID:  operatorID,
	}

	if err := tx.Create(&blacklist).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("创建黑名单记录失败", zap.Error(err))
		return errors.New("添加黑名单失败")
	}

	// 更新读者状态
	if err := tx.Model(&model.Reader{}).Where("id = ?", readerID).
		Updates(map[string]interface{}{
			"is_blacklisted": true,
			"status":         model.ReaderStatusInactive,
		}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新读者状态失败")
	}

	// 发送站内通知（失败不影响主流程）
	var reader model.Reader
	if err := tx.Where("id = ?", readerID).First(&reader).Error; err == nil && reader.UserID > 0 {
		title := "⚠️ 账户借阅权限已停用"
		content := fmt.Sprintf("由于「%s」，您的读者账户已被加入黑名单并停用借阅权限。", reason)
		if description != "" {
			content += " 原因说明：" + description
		}
		_ = (&MessageService{}).CreateMessage(ctx, reader.UserID, model.MessageTypeSystem, title, content, nil, "blacklist")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("添加黑名单失败")
	}

	global.GVA_LOG.Info("添加黑名单成功", zap.Uint("reader_id", readerID), zap.String("reason", string(reason)))
	return nil
}

// RemoveActiveBlacklistByUserID 通过用户ID解除该用户的所有生效黑名单
func (s *BlacklistService) RemoveActiveBlacklistByUserID(ctx context.Context, userID uint, operatorID uint, remark string) error {
	tx := global.DB(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var reader model.Reader
	if err := tx.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		tx.Rollback()
		return errors.New("读者信息不存在")
	}

	var activeCount int64
	if err := tx.Model(&model.Blacklist{}).
		Where("reader_id = ? AND status = ?", reader.ID, model.BlacklistStatusActive).
		Count(&activeCount).Error; err != nil {
		tx.Rollback()
		return errors.New("查询黑名单状态失败")
	}
	if activeCount == 0 {
		tx.Rollback()
		return errors.New("该用户当前不在黑名单中")
	}

	now := time.Now()
	updateRemark := remark
	if updateRemark == "" {
		updateRemark = "管理员手动解禁"
	}
	if err := tx.Model(&model.Blacklist{}).
		Where("reader_id = ? AND status = ?", reader.ID, model.BlacklistStatusActive).
		Updates(map[string]interface{}{
			"status":      model.BlacklistStatusLifted,
			"lifted_date": now,
			"operator_id": operatorID,
			"remark":      updateRemark,
		}).Error; err != nil {
		tx.Rollback()
		return errors.New("解除黑名单失败")
	}

	if err := tx.Model(&model.Reader{}).
		Where("id = ?", reader.ID).
		Updates(map[string]interface{}{
			"is_blacklisted": false,
			"status":         model.ReaderStatusActive,
		}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新读者状态失败")
	}

	// 发送站内通知（失败不影响主流程）
	_ = (&MessageService{}).CreateMessage(
		ctx,
		userID,
		model.MessageTypeSystem,
		"✅ 黑名单已解除",
		"管理员已解除您的黑名单限制，借阅权限已恢复。",
		nil,
		"blacklist",
	)

	if err := tx.Commit().Error; err != nil {
		return errors.New("解除黑名单失败")
	}
	return nil
}

// RemoveFromBlacklist 从黑名单移除（解禁）
func (s *BlacklistService) RemoveFromBlacklist(ctx context.Context, blacklistID uint, operatorID uint, remark string) error {
	tx := global.DB(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找黑名单记录
	var blacklist model.Blacklist
	if err := tx.First(&blacklist, blacklistID).Error; err != nil {
		tx.Rollback()
		return errors.New("黑名单记录不存在")
	}

	if blacklist.Status != model.BlacklistStatusActive {
		tx.Rollback()
		return errors.New("该黑名单已解除或过期")
	}

	// 更新黑名单状态
	now := time.Now()
	if err := tx.Model(&blacklist).Updates(map[string]interface{}{
		"status":      model.BlacklistStatusLifted,
		"lifted_date": now,
		"operator_id": operatorID,
		"remark":      remark,
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新黑名单记录失败")
	}

	// 检查该读者是否还有其他生效中的黑名单
	var activeCount int64
	tx.Model(&model.Blacklist{}).
		Where("reader_id = ? AND status = ? AND id != ?", blacklist.ReaderID, model.BlacklistStatusActive, blacklistID).
		Count(&activeCount)

	// 如果没有其他黑名单，解除读者的黑名单状态
	if activeCount == 0 {
		if err := tx.Model(&model.Reader{}).Where("id = ?", blacklist.ReaderID).
			Updates(map[string]interface{}{
				"is_blacklisted": false,
				"status":         model.ReaderStatusActive,
			}).Error; err != nil {
			tx.Rollback()
			return errors.New("更新读者状态失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("解除黑名单失败")
	}

	global.GVA_LOG.Info("解除黑名单成功", zap.Uint("blacklist_id", blacklistID), zap.Uint("reader_id", blacklist.ReaderID))
	return nil
}

// CheckAndAddOverdueBlacklist 检查并自动拉黑逾期严重的读者
// 定时任务调用
func (s *BlacklistService) CheckAndAddOverdueBlacklist() error {
	ctx := context.Background()
	// 获取配置
	blacklistDays := GlobalConfigService.GetIntConfig(ctx, model.ConfigOverdueBlacklistDays, 30)

	// 查找逾期超过指定天数的借阅记录
	var overdueRecords []model.BorrowRecord
	cutoffDate := time.Now().AddDate(0, 0, -blacklistDays)

	if err := global.DB(ctx).Where("status IN (?) AND due_date < ?",
		[]model.BorrowStatus{model.BorrowStatusBorrowed, model.BorrowStatusOverdue}, cutoffDate).
		Preload("Reader").
		Preload("Book").
		Find(&overdueRecords).Error; err != nil {
		return err
	}

	processedReaderIDs := make(map[uint]struct{}, len(overdueRecords))
	for _, record := range overdueRecords {
		// 同一轮任务内按 reader_id 去重，避免重复尝试拉黑同一读者
		if _, exists := processedReaderIDs[record.ReaderID]; exists {
			continue
		}
		processedReaderIDs[record.ReaderID] = struct{}{}

		// 检查读者是否已在黑名单
		if record.Reader.IsBlacklisted {
			continue
		}

		// 计算逾期天数
		overdueDays := int(time.Since(record.DueDate).Hours() / 24)

		// 添加到黑名单
		description := fmt.Sprintf("借阅《%s》逾期%d天未还", record.Book.Title, overdueDays)
		if err := s.AddToBlacklist(context.Background(), record.ReaderID, model.BlacklistReasonOverdue, description, nil, 0); err != nil {
			global.GVA_LOG.Error("自动拉黑失败", zap.Error(err), zap.Uint("reader_id", record.ReaderID))
			continue
		}

		global.GVA_LOG.Info("自动拉黑逾期读者", zap.Uint("reader_id", record.ReaderID), zap.Int("overdue_days", overdueDays))
	}

	return nil
}

// CheckExpiredBlacklist 检查并处理过期的黑名单
// 定时任务调用
func (s *BlacklistService) CheckExpiredBlacklist() error {
	ctx := context.Background()
	now := time.Now()

	// 查找所有过期的黑名单
	var expiredBlacklists []model.Blacklist
	if err := global.DB(ctx).Where("status = ? AND end_date IS NOT NULL AND end_date < ?",
		model.BlacklistStatusActive, now).
		Find(&expiredBlacklists).Error; err != nil {
		return err
	}

	if len(expiredBlacklists) == 0 {
		return nil
	}

	// 批量更新为expired
	blacklistIDs := make([]uint, len(expiredBlacklists))
	readerIDs := make(map[uint]bool)

	for i, bl := range expiredBlacklists {
		blacklistIDs[i] = bl.ID
		readerIDs[bl.ReaderID] = true
	}

	if err := global.DB(ctx).Model(&model.Blacklist{}).
		Where("id IN ?", blacklistIDs).
		Update("status", model.BlacklistStatusExpired).Error; err != nil {
		return err
	}

	// 对于每个读者，检查是否还有其他生效中的黑名单
	for readerID := range readerIDs {
		var activeCount int64
		global.DB(ctx).Model(&model.Blacklist{}).
			Where("reader_id = ? AND status = ?", readerID, model.BlacklistStatusActive).
			Count(&activeCount)

		// 如果没有其他黑名单，解除读者的黑名单状态
		if activeCount == 0 {
			global.DB(ctx).Model(&model.Reader{}).Where("id = ?", readerID).
				Updates(map[string]interface{}{
					"is_blacklisted": false,
					"status":         model.ReaderStatusActive,
				})
		}
	}

	global.GVA_LOG.Info("处理过期黑名单", zap.Int("count", len(expiredBlacklists)))
	return nil
}

// GetReaderBlacklists 获取读者的黑名单记录
func (s *BlacklistService) GetReaderBlacklists(ctx context.Context, readerID uint) ([]model.Blacklist, error) {
	var blacklists []model.Blacklist
	if err := global.DB(ctx).Where("reader_id = ?", readerID).
		Order("created_at DESC").
		Find(&blacklists).Error; err != nil {
		return nil, err
	}
	return blacklists, nil
}
