package initialize

import (
	"bookadmin/global"
	"bookadmin/observability"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
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

	// 分布式追踪：为 Redis 客户端注入 OpenTelemetry 埋点
	if observability.TracerEnabled() {
		if err := redisotel.InstrumentTracing(client); err != nil {
			global.GVA_LOG.Warn("Redis 追踪埋点失败", zap.Error(err))
		}
	}

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

// ensureConsumerGroup 确保消费者组存在：若 stream 不存在则创建 stream+group，若 group 已存在则跳过，避免 BUSYGROUP 产生错误 trace
func ensureConsumerGroup(ctx context.Context, client *redis.Client, stream, group string) error {
	groups, err := client.XInfoGroups(ctx, stream).Result()
	if err != nil {
		if strings.Contains(err.Error(), "no such key") || strings.Contains(err.Error(), "not exist") {
			return client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
		}
		return err
	}
	for _, g := range groups {
		if g.Name == group {
			return nil
		}
	}
	return client.XGroupCreate(ctx, stream, group, "0").Err()
}

// InitRedisStreamGroups 初始化Redis Stream消费者组
func InitRedisStreamGroups() {
	if global.GVA_REDIS == nil {
		global.GVA_LOG.Error("Redis未初始化，无法创建消费者组")
		return
	}

	ctx := context.Background()
	groupName := global.GVA_CONFIG.Worker.ConsumerGroup

	// 创建点赞操作Stream的消费者组
	likeStream := "stream:like:actions"
	if err := ensureConsumerGroup(ctx, global.GVA_REDIS, likeStream, groupName); err != nil {
		global.GVA_LOG.Error("创建点赞Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("点赞Stream消费者组创建成功: %s", likeStream))
	}

	// 创建收藏操作Stream的消费者组
	favoriteStream := "stream:favorite:actions"
	if err := ensureConsumerGroup(ctx, global.GVA_REDIS, favoriteStream, groupName); err != nil {
		global.GVA_LOG.Error("创建收藏Stream消费者组失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("收藏Stream消费者组创建成功: %s", favoriteStream))
	}

	// 创建死信Stream的消费者组
	if global.GVA_CONFIG.Worker.DeadLetterStream != "" {
		if err := ensureConsumerGroup(ctx, global.GVA_REDIS, global.GVA_CONFIG.Worker.DeadLetterStream, groupName); err != nil {
			global.GVA_LOG.Warn("创建死信Stream消费者组失败", zap.Error(err))
		}
	}
}
