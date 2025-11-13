package service

import (
	"bookadmin/constants"
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/utils"
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// RankingService 榜单服务
type RankingService struct {
	redis *RedisService
}

// NewRankingService 创建榜单服务实例
func NewRankingService() *RankingService {
	return &RankingService{
		redis: &RedisService{},
	}
}

// GetRanking 获取榜单
func (s *RankingService) GetRanking(ctx context.Context, rankingType model.RankingType, period model.RankingPeriod, limit int) (*model.RankingResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 100 // 默认返回Top100
	}

	// 1. 获取当前周期
	var periodKey string
	if period == model.RankingPeriodWeek {
		periodKey = utils.GetCurrentWeekPeriod()
	} else {
		periodKey = utils.GetCurrentMonthPeriod()
	}

	// 2. 构建Redis Key
	var redisKey string
	if rankingType == model.RankingTypeLike {
		if period == model.RankingPeriodWeek {
			redisKey = constants.KeyRankLikesWeek(periodKey)
		} else {
			redisKey = constants.KeyRankLikesMonth(periodKey)
		}
	} else {
		if period == model.RankingPeriodWeek {
			redisKey = constants.KeyRankFavoritesWeek(periodKey)
		} else {
			redisKey = constants.KeyRankFavoritesMonth(periodKey)
		}
	}

	// 3. 从Redis获取榜单数据
	rankData, err := s.redis.GetRankingTop(ctx, redisKey, limit)
	if err != nil {
		global.GVA_LOG.Error("获取榜单失败",
			zap.String("type", string(rankingType)),
			zap.String("period", string(period)),
			zap.Error(err))
		return nil, err
	}

	// 4. 如果Redis没有数据，从MySQL重建榜单
	if len(rankData) == 0 {
		global.GVA_LOG.Warn("榜单数据为空，尝试从MySQL重建",
			zap.String("type", string(rankingType)),
			zap.String("period", string(period)))

		if err := s.rebuildRanking(ctx, rankingType, period, periodKey); err != nil {
			global.GVA_LOG.Error("重建榜单失败", zap.Error(err))
			return &model.RankingResponse{
				Type:      rankingType,
				Period:    period,
				PeriodKey: periodKey,
				Items:     []model.RankingItem{},
				Total:     0,
				UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
			}, nil
		}

		// 重新获取
		rankData, _ = s.redis.GetRankingTop(ctx, redisKey, limit)
	}

	// 5. 构建榜单响应
	items := make([]model.RankingItem, 0, len(rankData))
	for i, data := range rankData {
		bookID, _ := strconv.ParseUint(data.Member.(string), 10, 32)

		// 查询图书详情
		var book model.Book
		if err := global.GVA_DB.First(&book, bookID).Error; err == nil {
			items = append(items, model.RankingItem{
				Rank:   i + 1,
				BookID: uint(bookID),
				Book:   &book,
				Score:  int64(data.Score),
			})
		}
	}

	return &model.RankingResponse{
		Type:      rankingType,
		Period:    period,
		PeriodKey: periodKey,
		Items:     items,
		Total:     len(items),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// rebuildRanking 重建榜单（从MySQL统计数据）
func (s *RankingService) rebuildRanking(ctx context.Context, rankingType model.RankingType, period model.RankingPeriod, periodKey string) error {
	global.GVA_LOG.Info("开始重建榜单",
		zap.String("type", string(rankingType)),
		zap.String("period", string(period)),
		zap.String("periodKey", periodKey))

	// 1. 获取时间范围
	var startTime, endTime time.Time
	var err error

	if period == model.RankingPeriodWeek {
		startTime, endTime, err = utils.GetWeekStartEnd(periodKey)
	} else {
		startTime, endTime, err = utils.GetMonthStartEnd(periodKey)
	}

	if err != nil {
		return err
	}

	// 2. 从MySQL统计数据
	type RankItem struct {
		BookID uint
		Count  int64
	}

	var items []RankItem

	if rankingType == model.RankingTypeLike {
		// 统计点赞数
		if err := global.GVA_DB.Model(&model.BookLike{}).
			Select("book_id, COUNT(*) as count").
			Where("created_at BETWEEN ? AND ?", startTime, endTime).
			Group("book_id").
			Order("count DESC").
			Limit(100).
			Scan(&items).Error; err != nil {
			return err
		}
	} else {
		// 统计收藏数
		if err := global.GVA_DB.Model(&model.BookFavorite{}).
			Select("book_id, COUNT(*) as count").
			Where("created_at BETWEEN ? AND ?", startTime, endTime).
			Group("book_id").
			Order("count DESC").
			Limit(100).
			Scan(&items).Error; err != nil {
			return err
		}
	}

	// 3. 写入Redis
	var redisKey string
	if rankingType == model.RankingTypeLike {
		if period == model.RankingPeriodWeek {
			redisKey = constants.KeyRankLikesWeek(periodKey)
		} else {
			redisKey = constants.KeyRankLikesMonth(periodKey)
		}
	} else {
		if period == model.RankingPeriodWeek {
			redisKey = constants.KeyRankFavoritesWeek(periodKey)
		} else {
			redisKey = constants.KeyRankFavoritesMonth(periodKey)
		}
	}

	// 清空旧数据
	global.GVA_REDIS.Del(ctx, redisKey)

	// 批量写入
	for _, item := range items {
		if err := s.redis.IncrRankingScore(ctx, redisKey, item.BookID, float64(item.Count)); err != nil {
			global.GVA_LOG.Warn("写入榜单数据失败",
				zap.Uint("bookID", item.BookID),
				zap.Error(err))
		}
	}

	// 设置过期时间
	var expireTime time.Duration
	if period == model.RankingPeriodWeek {
		expireTime = time.Duration(constants.ExpireWeekRank) * time.Second
	} else {
		expireTime = time.Duration(constants.ExpireMonthRank) * time.Second
	}
	global.GVA_REDIS.Expire(ctx, redisKey, expireTime)

	global.GVA_LOG.Info("榜单重建完成",
		zap.String("type", string(rankingType)),
		zap.String("period", string(period)),
		zap.Int("count", len(items)))

	return nil
}

// RebuildAllRankings 重建所有榜单（定时任务调用）
func (s *RankingService) RebuildAllRankings(ctx context.Context) error {
	global.GVA_LOG.Info("开始重建所有榜单")

	// 当前周期
	weekPeriod := utils.GetCurrentWeekPeriod()
	monthPeriod := utils.GetCurrentMonthPeriod()

	// 重建4个榜单
	rankings := []struct {
		rankingType model.RankingType
		period      model.RankingPeriod
		periodKey   string
	}{
		{model.RankingTypeLike, model.RankingPeriodWeek, weekPeriod},
		{model.RankingTypeLike, model.RankingPeriodMonth, monthPeriod},
		{model.RankingTypeFavorite, model.RankingPeriodWeek, weekPeriod},
		{model.RankingTypeFavorite, model.RankingPeriodMonth, monthPeriod},
	}

	for _, r := range rankings {
		if err := s.rebuildRanking(ctx, r.rankingType, r.period, r.periodKey); err != nil {
			global.GVA_LOG.Error("重建榜单失败",
				zap.String("type", string(r.rankingType)),
				zap.String("period", string(r.period)),
				zap.Error(err))
		}
	}

	global.GVA_LOG.Info("所有榜单重建完成")
	return nil
}

// SyncBookStatsToRedis 同步图书统计数据到Redis（用于缓存预热）
func (s *RankingService) SyncBookStatsToRedis(ctx context.Context) error {
	global.GVA_LOG.Info("开始同步图书统计数据到Redis")

	// 查询所有图书的统计数据
	var books []model.Book
	if err := global.GVA_DB.Select("id, like_count, favorite_count").Find(&books).Error; err != nil {
		return err
	}

	// 批量写入Redis
	for _, book := range books {
		if err := s.redis.SetBookStats(ctx, book.ID, int(book.LikeCount), int(book.FavoriteCount)); err != nil {
			global.GVA_LOG.Warn("同步图书统计失败",
				zap.Uint("bookID", book.ID),
				zap.Error(err))
		}
	}

	global.GVA_LOG.Info(fmt.Sprintf("图书统计数据同步完成，共 %d 本书", len(books)))
	return nil
}
