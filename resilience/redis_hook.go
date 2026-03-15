package resilience

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// redisBreakerHook 实现 redis.Hook，将所有 Redis 命令通过熔断器执行
type redisBreakerHook struct{}

// NewRedisBreakerHook 创建 Redis 熔断器 Hook
func NewRedisBreakerHook() *redisBreakerHook {
	return &redisBreakerHook{}
}

// DialHook 透传
func (h *redisBreakerHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook 将命令执行包裹在熔断器中
func (h *redisBreakerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		_, err := RedisExecute(func() (interface{}, error) {
			return nil, next(ctx, cmd)
		})
		return err
	}
}

// ProcessPipelineHook 将 Pipeline 执行包裹在熔断器中
func (h *redisBreakerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		_, err := RedisExecute(func() (interface{}, error) {
			return nil, next(ctx, cmds)
		})
		return err
	}
}
