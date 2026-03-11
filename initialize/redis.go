package initialize

import (
	"bookadmin/global"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis 初始化Redis连接
func Redis() *redis.Client {
	if global.GVA_REDIS != nil {
		return global.GVA_REDIS
	}

	if global.GVA_CONFIG == nil {
		global.GVA_LOG.Error("配置未初始化，无法连接Redis")
		return nil
	}
	if !global.GVA_CONFIG.Redis.Enabled {
		global.GVA_LOG.Warn("Redis已在配置中禁用")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         global.GVA_CONFIG.Redis.Addr,
		Password:     global.GVA_CONFIG.Redis.Password,
		DB:           global.GVA_CONFIG.Redis.DB,
		PoolSize:     global.GVA_CONFIG.Redis.PoolSize,
		MinIdleConns: global.GVA_CONFIG.Redis.MinIdleConns,
		MaxRetries:   global.GVA_CONFIG.Redis.MaxRetries,
		DialTimeout:  time.Duration(global.GVA_CONFIG.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(global.GVA_CONFIG.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(global.GVA_CONFIG.Redis.WriteTimeout) * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		global.GVA_LOG.Error("Redis连接失败", zap.Error(err))
		return nil
	}

	global.GVA_LOG.Info("Redis连接成功")
	global.GVA_REDIS = client
	return client
}

// InitRedisStreamGroups 初始化Redis Stream消费者组
func InitRedisStreamGroups() {
	if global.GVA_REDIS == nil {
		global.GVA_LOG.Error("Redis未初始化，无法创建消费者组")
		return
	}

	ctx := context.Background()

	// 创建点赞操作Stream的消费者组
	likeStream := "stream:like:actions"
	groupName := global.GVA_CONFIG.Worker.ConsumerGroup
	err := global.GVA_REDIS.XGroupCreateMkStream(ctx, likeStream, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		global.GVA_LOG.Error("创建点赞Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("点赞Stream消费者组创建成功: %s", likeStream))
	}

	// 创建收藏操作Stream的消费者组
	favoriteStream := "stream:favorite:actions"
	err = global.GVA_REDIS.XGroupCreateMkStream(ctx, favoriteStream, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		global.GVA_LOG.Error("创建收藏Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("收藏Stream消费者组创建成功: %s", favoriteStream))
	}

	if global.GVA_CONFIG.Worker.DeadLetterStream != "" {
		if err := global.GVA_REDIS.XGroupCreateMkStream(ctx, global.GVA_CONFIG.Worker.DeadLetterStream, groupName, "0").Err(); err != nil &&
			err.Error() != "BUSYGROUP Consumer Group name already exists" {
			global.GVA_LOG.Warn("创建死信Stream消费者组失败", zap.Error(err))
		}
	}
}
