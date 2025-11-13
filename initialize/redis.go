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
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379", // Redis地址
		Password:     "",               // Redis密码
		DB:           0,                // Redis数据库
		PoolSize:     100,              // 连接池大小
		MinIdleConns: 10,               // 最小空闲连接数
		MaxRetries:   3,                // 最大重试次数
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读取超时
		WriteTimeout: 3 * time.Second,  // 写入超时
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
	err := global.GVA_REDIS.XGroupCreateMkStream(ctx, likeStream, "sync-group", "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		global.GVA_LOG.Error("创建点赞Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("点赞Stream消费者组创建成功: %s", likeStream))
	}

	// 创建收藏操作Stream的消费者组
	favoriteStream := "stream:favorite:actions"
	err = global.GVA_REDIS.XGroupCreateMkStream(ctx, favoriteStream, "sync-group", "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		global.GVA_LOG.Error("创建收藏Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("收藏Stream消费者组创建成功: %s", favoriteStream))
	}
}
