package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"context"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type MessageService struct{}

// CreateMessage 创建消息
func (s *MessageService) CreateMessage(ctx context.Context, userID uint, msgType model.MessageType, title, content string, relatedID *uint, relatedType string) error {
	message := model.Message{
		UserID:      userID,
		Type:        msgType,
		Title:       title,
		Content:     content,
		IsRead:      false,
		RelatedID:   relatedID,
		RelatedType: relatedType,
	}

	if err := global.DB(ctx).Create(&message).Error; err != nil {
		global.GVA_LOG.Error("创建消息失败", zap.Error(err))
		return errors.New("创建消息失败")
	}

	global.GVA_LOG.Info("创建消息成功", zap.Uint("user_id", userID), zap.String("title", title))
	return nil
}

// GetUserMessages 获取用户消息列表
func (s *MessageService) GetUserMessages(ctx context.Context, userID uint, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	db := global.DB(ctx).Model(&model.Message{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按创建时间倒序
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetUnreadCount 获取未读消息数量
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := global.DB(ctx).Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(ctx context.Context, messageID, userID uint) error {
	now := time.Now()
	result := global.DB(ctx).Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("消息不存在或无权操作")
	}

	return nil
}

// MarkAllAsRead 标记所有消息为已读
func (s *MessageService) MarkAllAsRead(ctx context.Context, userID uint) error {
	now := time.Now()
	return global.DB(ctx).Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, messageID, userID uint) error {
	result := global.DB(ctx).Where("id = ? AND user_id = ?", messageID, userID).
		Delete(&model.Message{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("消息不存在或无权操作")
	}

	return nil
}

// SendReservationAvailableMessage 发送预约可取书消息
func (s *MessageService) SendReservationAvailableMessage(ctx context.Context, userID uint, bookTitle string, reservationID uint, pickupDays int) error {
	title := "📚 预约图书已可借阅"
	content := "您预约的《" + bookTitle + "》现在可以借阅了！请在 " +
		strconv.Itoa(pickupDays) + " 天内前往图书管理员处登记借书。逾期预约将自动取消。"

	relatedID := reservationID
	return s.CreateMessage(ctx, userID, model.MessageTypeReservation, title, content, &relatedID, "reservation")
}
