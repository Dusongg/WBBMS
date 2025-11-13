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

// FavoriteService 收藏服务
type FavoriteService struct {
	redis *RedisService
}

// NewFavoriteService 创建收藏服务实例
func NewFavoriteService() *FavoriteService {
	return &FavoriteService{
		redis: &RedisService{},
	}
}

// ToggleFavorite 切换收藏状态（收藏/取消收藏）
func (s *FavoriteService) ToggleFavorite(ctx context.Context, userID, bookID uint) (*model.FavoriteStatus, error) {
	// 1. 防重复操作：加锁
	lockKey := constants.KeyFavoriteLock(userID, bookID)
	locked, err := s.redis.TryLock(ctx, lockKey, time.Second)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("操作太频繁，请稍后再试")
	}
	defer s.redis.Unlock(ctx, lockKey)

	// 2. 检查当前收藏状态
	isFavorited, err := s.checkFavoriteStatusFromRedis(ctx, userID, bookID)
	if err != nil {
		// Redis查询失败，从MySQL查询
		isFavorited, err = s.checkFavoriteStatusFromDB(userID, bookID)
		if err != nil {
			global.GVA_LOG.Error("查询收藏状态失败", zap.Error(err))
			return nil, errors.New("查询收藏状态失败")
		}
	}

	// 3. 执行切换操作
	var action string
	var delta int64

	if isFavorited {
		// 取消收藏
		action = "unfavorite"
		delta = -1
		if err := s.unfavoriteInRedis(ctx, userID, bookID); err != nil {
			global.GVA_LOG.Error("取消收藏失败", zap.Error(err))
			return nil, errors.New("取消收藏失败")
		}
	} else {
		// 收藏
		action = "favorite"
		delta = 1
		if err := s.favoriteInRedis(ctx, userID, bookID); err != nil {
			global.GVA_LOG.Error("收藏失败", zap.Error(err))
			return nil, errors.New("收藏失败")
		}
	}

	// 4. 更新榜单
	if err := s.updateRankings(ctx, bookID, float64(delta)); err != nil {
		global.GVA_LOG.Warn("更新榜单失败", zap.Error(err))
	}

	// 5. 发送消息到Stream异步同步MySQL
	if err := s.sendToStream(ctx, userID, bookID, action); err != nil {
		global.GVA_LOG.Warn("发送Stream消息失败", zap.Error(err))
		// Stream发送失败，回退Redis操作并直接同步MySQL
		s.rollbackRedis(ctx, userID, bookID, action)
		s.syncToMySQLDirectly(userID, bookID, action)
	}

	// 6. 获取最新统计数据
	_, favoriteCount, err := s.redis.GetBookStats(ctx, bookID)
	if err != nil {
		favoriteCount = s.getBookFavoriteCountFromDB(bookID)
	}

	return &model.FavoriteStatus{
		IsFavorited:   !isFavorited,
		FavoriteCount: favoriteCount,
	}, nil
}

// GetFavoriteStatus 查询收藏状态
func (s *FavoriteService) GetFavoriteStatus(ctx context.Context, userID, bookID uint) (*model.FavoriteStatus, error) {
	// 1. 先从Redis查询
	isFavorited, err := s.checkFavoriteStatusFromRedis(ctx, userID, bookID)
	if err != nil {
		// Redis查询失败，从MySQL查询
		isFavorited, err = s.checkFavoriteStatusFromDB(userID, bookID)
		if err != nil {
			return nil, err
		}
	}

	// 2. 获取图书收藏总数
	_, favoriteCount, err := s.redis.GetBookStats(ctx, bookID)
	if err != nil {
		favoriteCount = s.getBookFavoriteCountFromDB(bookID)
	}

	return &model.FavoriteStatus{
		IsFavorited:   isFavorited,
		FavoriteCount: favoriteCount,
	}, nil
}

// BatchGetFavoriteStatus 批量获取收藏状态
func (s *FavoriteService) BatchGetFavoriteStatus(ctx context.Context, userID uint, bookIDs []uint) ([]model.BatchFavoriteStatus, error) {
	if len(bookIDs) == 0 {
		return []model.BatchFavoriteStatus{}, nil
	}

	// 批量查询用户收藏状态
	favoriteMap, err := s.redis.BatchCheckUserFavorites(ctx, userID, bookIDs)
	if err != nil {
		global.GVA_LOG.Warn("批量查询收藏状态失败", zap.Error(err))
		favoriteMap = make(map[uint]bool)
	}

	// 构建结果
	result := make([]model.BatchFavoriteStatus, 0, len(bookIDs))
	for _, bookID := range bookIDs {
		_, favoriteCount, _ := s.redis.GetBookStats(ctx, bookID)
		result = append(result, model.BatchFavoriteStatus{
			BookID:        bookID,
			IsFavorited:   favoriteMap[bookID],
			FavoriteCount: favoriteCount,
		})
	}

	return result, nil
}

// GetUserFavoriteList 获取用户收藏列表
func (s *FavoriteService) GetUserFavoriteList(ctx context.Context, userID uint, page, pageSize int) ([]model.BookFavoriteWithBook, int64, error) {
	var favorites []model.BookFavoriteWithBook
	var total int64

	db := global.GVA_DB.Model(&model.BookFavorite{}).Where("user_id = ?", userID)

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
		Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

// ============================================
// 内部辅助方法
// ============================================

func (s *FavoriteService) checkFavoriteStatusFromRedis(ctx context.Context, userID, bookID uint) (bool, error) {
	return s.redis.IsUserFavorited(ctx, userID, bookID)
}

func (s *FavoriteService) checkFavoriteStatusFromDB(userID, bookID uint) (bool, error) {
	var count int64
	err := global.GVA_DB.Model(&model.BookFavorite{}).
		Where("user_id = ? AND book_id = ?", userID, bookID).
		Count(&count).Error
	return count > 0, err
}

func (s *FavoriteService) favoriteInRedis(ctx context.Context, userID, bookID uint) error {
	if err := s.redis.AddUserFavorite(ctx, userID, bookID); err != nil {
		return err
	}
	if err := s.redis.IncrBookFavoriteCount(ctx, bookID, 1); err != nil {
		return err
	}
	return nil
}

func (s *FavoriteService) unfavoriteInRedis(ctx context.Context, userID, bookID uint) error {
	if err := s.redis.RemoveUserFavorite(ctx, userID, bookID); err != nil {
		return err
	}
	if err := s.redis.IncrBookFavoriteCount(ctx, bookID, -1); err != nil {
		return err
	}
	return nil
}

func (s *FavoriteService) updateRankings(ctx context.Context, bookID uint, delta float64) error {
	now := time.Now()

	// 周榜
	year, week := now.ISOWeek()
	weekPeriod := fmt.Sprintf("%d-W%02d", year, week)
	weekKey := constants.KeyRankFavoritesWeek(weekPeriod)
	if err := s.redis.IncrRankingScore(ctx, weekKey, bookID, delta); err != nil {
		return err
	}

	// 月榜
	monthPeriod := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	monthKey := constants.KeyRankFavoritesMonth(monthPeriod)
	if err := s.redis.IncrRankingScore(ctx, monthKey, bookID, delta); err != nil {
		return err
	}

	return nil
}

func (s *FavoriteService) sendToStream(ctx context.Context, userID, bookID uint, action string) error {
	values := map[string]interface{}{
		"user_id":   userID,
		"book_id":   bookID,
		"action":    action, // "favorite" or "unfavorite"
		"timestamp": time.Now().Unix(),
	}
	return s.redis.AddToStream(ctx, constants.StreamFavoriteActions, values)
}

func (s *FavoriteService) rollbackRedis(ctx context.Context, userID, bookID uint, action string) {
	if action == "favorite" {
		s.unfavoriteInRedis(ctx, userID, bookID)
		s.updateRankings(ctx, bookID, -1)
	} else {
		s.favoriteInRedis(ctx, userID, bookID)
		s.updateRankings(ctx, bookID, 1)
	}
}

func (s *FavoriteService) syncToMySQLDirectly(userID, bookID uint, action string) {
	if action == "favorite" {
		favorite := model.BookFavorite{
			UserID: userID,
			BookID: bookID,
		}
		if err := global.GVA_DB.Create(&favorite).Error; err != nil {
			global.GVA_LOG.Error("直接同步收藏到MySQL失败", zap.Error(err))
			return
		}
		global.GVA_DB.Model(&model.Book{}).
			Where("id = ?", bookID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1))
	} else {
		if err := global.GVA_DB.Where("user_id = ? AND book_id = ?", userID, bookID).
			Delete(&model.BookFavorite{}).Error; err != nil {
			global.GVA_LOG.Error("直接同步取消收藏到MySQL失败", zap.Error(err))
			return
		}
		global.GVA_DB.Model(&model.Book{}).
			Where("id = ?", bookID).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1))
	}
}

func (s *FavoriteService) getBookFavoriteCountFromDB(bookID uint) int {
	var count int64
	global.GVA_DB.Model(&model.BookFavorite{}).Where("book_id = ?", bookID).Count(&count)
	return int(count)
}
