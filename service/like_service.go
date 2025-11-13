package service

import (
	"bookadmin/constants"
	"bookadmin/global"
	"bookadmin/model"
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LikeService 点赞服务
type LikeService struct {
	redis *RedisService
}

// NewLikeService 创建点赞服务实例
func NewLikeService() *LikeService {
	return &LikeService{
		redis: &RedisService{},
	}
}

// ToggleLike 切换点赞状态（点赞/取消点赞）
// 这是最核心的功能，采用高性能异步方案：
// 1. 先操作Redis（快速返回）
// 2. 发送消息到Stream异步同步MySQL
func (s *LikeService) ToggleLike(ctx context.Context, userID, bookID uint) (*model.LikeStatus, error) {
	// 1. 防重复操作：加锁（1秒内同一用户对同一本书只能操作一次）
	lockKey := constants.KeyLikeLock(userID, bookID)
	locked, err := s.redis.TryLock(ctx, lockKey, time.Second)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("操作太频繁，请稍后再试")
	}
	defer s.redis.Unlock(ctx, lockKey)

	// 2. 检查当前点赞状态
	isLiked, err := s.checkLikeStatusFromRedis(ctx, userID, bookID)
	if err != nil {
		// Redis查询失败，从MySQL查询
		isLiked, err = s.checkLikeStatusFromDB(userID, bookID)
		if err != nil {
			global.GVA_LOG.Error("查询点赞状态失败", zap.Error(err))
			return nil, errors.New("查询点赞状态失败")
		}
	}

	// 3. 执行切换操作
	var action string
	var delta int64

	if isLiked {
		// 取消点赞
		action = "unlike"
		delta = -1
		if err := s.unlikeInRedis(ctx, userID, bookID); err != nil {
			global.GVA_LOG.Error("取消点赞失败", zap.Error(err))
			return nil, errors.New("取消点赞失败")
		}
	} else {
		// 点赞
		action = "like"
		delta = 1
		if err := s.likeInRedis(ctx, userID, bookID); err != nil {
			global.GVA_LOG.Error("点赞失败", zap.Error(err))
			return nil, errors.New("点赞失败")
		}
	}

	// 4. 更新榜单（周榜和月榜）
	if err := s.updateRankings(ctx, bookID, float64(delta)); err != nil {
		global.GVA_LOG.Warn("更新榜单失败", zap.Error(err))
		// 榜单更新失败不影响点赞操作
	}

	// 5. 发送消息到Stream异步同步MySQL
	if err := s.sendToStream(ctx, userID, bookID, action); err != nil {
		global.GVA_LOG.Warn("发送Stream消息失败", zap.Error(err))
		// Stream发送失败，回退Redis操作并直接同步MySQL
		s.rollbackRedis(ctx, userID, bookID, action)
		s.syncToMySQLDirectly(userID, bookID, action)
	}

	// 6. 获取最新统计数据
	likeCount, _, err := s.redis.GetBookStats(ctx, bookID)
	if err != nil {
		// Redis失败，从MySQL查询
		likeCount = s.getBookLikeCountFromDB(bookID)
	}

	return &model.LikeStatus{
		IsLiked:   !isLiked, // 返回切换后的状态
		LikeCount: likeCount,
	}, nil
}

// GetLikeStatus 查询点赞状态
func (s *LikeService) GetLikeStatus(ctx context.Context, userID, bookID uint) (*model.LikeStatus, error) {
	// 1. 先从Redis查询用户是否已点赞
	isLiked, err := s.checkLikeStatusFromRedis(ctx, userID, bookID)
	if err != nil {
		// Redis查询失败，从MySQL查询
		isLiked, err = s.checkLikeStatusFromDB(userID, bookID)
		if err != nil {
			return nil, err
		}
	}

	// 2. 获取图书点赞总数
	likeCount, _, err := s.redis.GetBookStats(ctx, bookID)
	if err != nil {
		// Redis失败，从MySQL查询
		likeCount = s.getBookLikeCountFromDB(bookID)
	}

	return &model.LikeStatus{
		IsLiked:   isLiked,
		LikeCount: likeCount,
	}, nil
}

// BatchGetLikeStatus 批量获取点赞状态（用于图书列表）
func (s *LikeService) BatchGetLikeStatus(ctx context.Context, userID uint, bookIDs []uint) ([]model.BatchLikeStatus, error) {
	if len(bookIDs) == 0 {
		return []model.BatchLikeStatus{}, nil
	}

	// 批量查询用户点赞状态
	likeMap, err := s.redis.BatchCheckUserLikes(ctx, userID, bookIDs)
	if err != nil {
		global.GVA_LOG.Warn("批量查询点赞状态失败", zap.Error(err))
		likeMap = make(map[uint]bool)
	}

	// 构建结果
	result := make([]model.BatchLikeStatus, 0, len(bookIDs))
	for _, bookID := range bookIDs {
		likeCount, _, _ := s.redis.GetBookStats(ctx, bookID)
		result = append(result, model.BatchLikeStatus{
			BookID:    bookID,
			IsLiked:   likeMap[bookID],
			LikeCount: likeCount,
		})
	}

	return result, nil
}

// GetUserLikeList 获取用户点赞列表
func (s *LikeService) GetUserLikeList(ctx context.Context, userID uint, page, pageSize int) ([]model.BookLikeWithBook, int64, error) {
	var likes []model.BookLikeWithBook
	var total int64

	db := global.GVA_DB.Model(&model.BookLike{}).Where("user_id = ?", userID)

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载图书信息
	offset := (page - 1) * pageSize
	if err := db.Preload("Book").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	return likes, total, nil
}

// ============================================
// 内部辅助方法
// ============================================

// checkLikeStatusFromRedis 从Redis检查点赞状态
func (s *LikeService) checkLikeStatusFromRedis(ctx context.Context, userID, bookID uint) (bool, error) {
	return s.redis.IsUserLiked(ctx, userID, bookID)
}

// checkLikeStatusFromDB 从MySQL检查点赞状态
func (s *LikeService) checkLikeStatusFromDB(userID, bookID uint) (bool, error) {
	var count int64
	err := global.GVA_DB.Model(&model.BookLike{}).
		Where("user_id = ? AND book_id = ?", userID, bookID).
		Count(&count).Error
	return count > 0, err
}

// likeInRedis 在Redis中执行点赞操作
func (s *LikeService) likeInRedis(ctx context.Context, userID, bookID uint) error {
	// 1. 添加用户点赞状态
	if err := s.redis.AddUserLike(ctx, userID, bookID); err != nil {
		return err
	}

	// 2. 增加图书点赞数
	if err := s.redis.IncrBookLikeCount(ctx, bookID, 1); err != nil {
		return err
	}

	return nil
}

// unlikeInRedis 在Redis中执行取消点赞操作
func (s *LikeService) unlikeInRedis(ctx context.Context, userID, bookID uint) error {
	// 1. 移除用户点赞状态
	if err := s.redis.RemoveUserLike(ctx, userID, bookID); err != nil {
		return err
	}

	// 2. 减少图书点赞数
	if err := s.redis.IncrBookLikeCount(ctx, bookID, -1); err != nil {
		return err
	}

	return nil
}

// updateRankings 更新榜单
func (s *LikeService) updateRankings(ctx context.Context, bookID uint, delta float64) error {
	now := time.Now()

	// 周榜：2025-W45
	year, week := now.ISOWeek()
	weekPeriod := fmt.Sprintf("%d-W%02d", year, week)
	weekKey := constants.KeyRankLikesWeek(weekPeriod)
	if err := s.redis.IncrRankingScore(ctx, weekKey, bookID, delta); err != nil {
		return err
	}

	// 月榜：2025-11
	monthPeriod := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	monthKey := constants.KeyRankLikesMonth(monthPeriod)
	if err := s.redis.IncrRankingScore(ctx, monthKey, bookID, delta); err != nil {
		return err
	}

	return nil
}

// sendToStream 发送操作到Stream异步队列
func (s *LikeService) sendToStream(ctx context.Context, userID, bookID uint, action string) error {
	values := map[string]interface{}{
		"user_id":   userID,
		"book_id":   bookID,
		"action":    action, // "like" or "unlike"
		"timestamp": time.Now().Unix(),
	}
	return s.redis.AddToStream(ctx, constants.StreamLikeActions, values)
}

// rollbackRedis 回滚Redis操作（Stream失败时）
func (s *LikeService) rollbackRedis(ctx context.Context, userID, bookID uint, action string) {
	if action == "like" {
		s.unlikeInRedis(ctx, userID, bookID)
		s.updateRankings(ctx, bookID, -1)
	} else {
		s.likeInRedis(ctx, userID, bookID)
		s.updateRankings(ctx, bookID, 1)
	}
}

// syncToMySQLDirectly 直接同步到MySQL（降级方案）
func (s *LikeService) syncToMySQLDirectly(userID, bookID uint, action string) {
	if action == "like" {
		// 插入点赞记录
		like := model.BookLike{
			UserID: userID,
			BookID: bookID,
		}
		if err := global.GVA_DB.Create(&like).Error; err != nil {
			global.GVA_LOG.Error("直接同步点赞到MySQL失败", zap.Error(err))
			return
		}
		// 更新books表统计
		global.GVA_DB.Model(&model.Book{}).
			Where("id = ?", bookID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))
	} else {
		// 删除点赞记录
		if err := global.GVA_DB.Where("user_id = ? AND book_id = ?", userID, bookID).
			Delete(&model.BookLike{}).Error; err != nil {
			global.GVA_LOG.Error("直接同步取消点赞到MySQL失败", zap.Error(err))
			return
		}
		// 更新books表统计
		global.GVA_DB.Model(&model.Book{}).
			Where("id = ?", bookID).
			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))
	}
}

// getBookLikeCountFromDB 从MySQL获取图书点赞数
func (s *LikeService) getBookLikeCountFromDB(bookID uint) int {
	var count int64
	global.GVA_DB.Model(&model.BookLike{}).Where("book_id = ?", bookID).Count(&count)
	return int(count)
}
