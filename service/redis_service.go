package service

import (
	"bookadmin/constants"
	"bookadmin/global"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService Redis操作服务
type RedisService struct{}

// ============================================
// 用户点赞/收藏状态操作
// ============================================

// AddUserLike 添加用户点赞状态到Redis
func (r *RedisService) AddUserLike(ctx context.Context, userID, bookID uint) error {
	key := constants.KeyUserLikes(userID)
	return global.GVA_REDIS.SAdd(ctx, key, bookID).Err()
}

// RemoveUserLike 移除用户点赞状态
func (r *RedisService) RemoveUserLike(ctx context.Context, userID, bookID uint) error {
	key := constants.KeyUserLikes(userID)
	return global.GVA_REDIS.SRem(ctx, key, bookID).Err()
}

// IsUserLiked 检查用户是否已点赞
func (r *RedisService) IsUserLiked(ctx context.Context, userID, bookID uint) (bool, error) {
	key := constants.KeyUserLikes(userID)
	return global.GVA_REDIS.SIsMember(ctx, key, bookID).Result()
}

// AddUserFavorite 添加用户收藏状态到Redis
func (r *RedisService) AddUserFavorite(ctx context.Context, userID, bookID uint) error {
	key := constants.KeyUserFavorites(userID)
	return global.GVA_REDIS.SAdd(ctx, key, bookID).Err()
}

// RemoveUserFavorite 移除用户收藏状态
func (r *RedisService) RemoveUserFavorite(ctx context.Context, userID, bookID uint) error {
	key := constants.KeyUserFavorites(userID)
	return global.GVA_REDIS.SRem(ctx, key, bookID).Err()
}

// IsUserFavorited 检查用户是否已收藏
func (r *RedisService) IsUserFavorited(ctx context.Context, userID, bookID uint) (bool, error) {
	key := constants.KeyUserFavorites(userID)
	return global.GVA_REDIS.SIsMember(ctx, key, bookID).Result()
}

// ============================================
// 图书统计信息操作
// ============================================

// IncrBookLikeCount 增加图书点赞数
func (r *RedisService) IncrBookLikeCount(ctx context.Context, bookID uint, delta int64) error {
	key := constants.KeyBookStats(bookID)
	return global.GVA_REDIS.HIncrBy(ctx, key, "like_count", delta).Err()
}

// IncrBookFavoriteCount 增加图书收藏数
func (r *RedisService) IncrBookFavoriteCount(ctx context.Context, bookID uint, delta int64) error {
	key := constants.KeyBookStats(bookID)
	return global.GVA_REDIS.HIncrBy(ctx, key, "favorite_count", delta).Err()
}

// GetBookStats 获取图书统计信息
func (r *RedisService) GetBookStats(ctx context.Context, bookID uint) (likeCount, favoriteCount int, err error) {
	key := constants.KeyBookStats(bookID)
	result, err := global.GVA_REDIS.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}

	if likeStr, ok := result["like_count"]; ok {
		likeCount, _ = strconv.Atoi(likeStr)
	}

	if favoriteStr, ok := result["favorite_count"]; ok {
		favoriteCount, _ = strconv.Atoi(favoriteStr)
	}

	return likeCount, favoriteCount, nil
}

// SetBookStats 设置图书统计信息
func (r *RedisService) SetBookStats(ctx context.Context, bookID uint, likeCount, favoriteCount int) error {
	key := constants.KeyBookStats(bookID)
	pipe := global.GVA_REDIS.Pipeline()
	pipe.HSet(ctx, key, "like_count", likeCount)
	pipe.HSet(ctx, key, "favorite_count", favoriteCount)
	pipe.Expire(ctx, key, time.Duration(constants.ExpireBookStats)*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}

// ============================================
// 榜单操作 (ZSet)
// ============================================

// IncrRankingScore 增加榜单分数
func (r *RedisService) IncrRankingScore(ctx context.Context, key string, bookID uint, delta float64) error {
	return global.GVA_REDIS.ZIncrBy(ctx, key, delta, fmt.Sprintf("%d", bookID)).Err()
}

// GetRankingTop 获取榜单前N名
func (r *RedisService) GetRankingTop(ctx context.Context, key string, limit int) ([]redis.Z, error) {
	return global.GVA_REDIS.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
}

// ============================================
// Stream 消息队列操作
// ============================================

// AddToStream 添加消息到Stream
func (r *RedisService) AddToStream(ctx context.Context, stream string, values map[string]interface{}) error {
	return global.GVA_REDIS.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// ============================================
// 分布式锁操作
// ============================================

// TryLock 尝试获取分布式锁
func (r *RedisService) TryLock(ctx context.Context, key string, duration time.Duration) (bool, error) {
	return global.GVA_REDIS.SetNX(ctx, key, "1", duration).Result()
}

// Unlock 释放分布式锁
func (r *RedisService) Unlock(ctx context.Context, key string) error {
	return global.GVA_REDIS.Del(ctx, key).Err()
}

// ============================================
// 批量操作
// ============================================

// BatchCheckUserLikes 批量检查用户点赞状态
func (r *RedisService) BatchCheckUserLikes(ctx context.Context, userID uint, bookIDs []uint) (map[uint]bool, error) {
	key := constants.KeyUserLikes(userID)
	result := make(map[uint]bool)

	// 将uint切片转换为interface切片
	members := make([]interface{}, len(bookIDs))
	for i, id := range bookIDs {
		members[i] = id
	}

	// 使用SMISMEMBER批量检查（Redis 6.2+）
	exists, err := global.GVA_REDIS.SMIsMember(ctx, key, members...).Result()
	if err != nil {
		return nil, err
	}

	for i, bookID := range bookIDs {
		result[bookID] = exists[i]
	}

	return result, nil
}

// BatchCheckUserFavorites 批量检查用户收藏状态
func (r *RedisService) BatchCheckUserFavorites(ctx context.Context, userID uint, bookIDs []uint) (map[uint]bool, error) {
	key := constants.KeyUserFavorites(userID)
	result := make(map[uint]bool)

	members := make([]interface{}, len(bookIDs))
	for i, id := range bookIDs {
		members[i] = id
	}

	exists, err := global.GVA_REDIS.SMIsMember(ctx, key, members...).Result()
	if err != nil {
		return nil, err
	}

	for i, bookID := range bookIDs {
		result[bookID] = exists[i]
	}

	return result, nil
}
