package resilience

import (
	"bookadmin/global"
	"time"

	"github.com/sony/gobreaker"
)

var (
	dbBreaker    *gobreaker.CircuitBreaker
	redisBreaker *gobreaker.CircuitBreaker
)

func initBreakers() {
	if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Resilience.CircuitBreaker.Enabled {
		return
	}
	cfg := global.GVA_CONFIG.Resilience.CircuitBreaker

	dbBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "mysql",
		MaxRequests: uint32(cfg.MaxRequests),
		Interval:    time.Duration(cfg.Interval) * time.Second,
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.FailureThreshold) ||
				counts.TotalFailures >= uint32(cfg.FailureThreshold)
		},
	})

	redisBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis",
		MaxRequests: uint32(cfg.MaxRequests),
		Interval:    time.Duration(cfg.Interval) * time.Second,
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.FailureThreshold) ||
				counts.TotalFailures >= uint32(cfg.FailureThreshold)
		},
	})
}

// InitBreakers 初始化熔断器（在配置加载后调用）
func InitBreakers() {
	initBreakers()
}

// DBExecute 通过熔断器执行 DB 操作
func DBExecute(fn func() (interface{}, error)) (interface{}, error) {
	if dbBreaker == nil {
		return fn()
	}
	return dbBreaker.Execute(fn)
}

// RedisExecute 通过熔断器执行 Redis 操作
func RedisExecute(fn func() (interface{}, error)) (interface{}, error) {
	if redisBreaker == nil {
		return fn()
	}
	return redisBreaker.Execute(fn)
}

// DBState 获取 DB 熔断器状态
func DBState() gobreaker.State {
	if dbBreaker == nil {
		return gobreaker.StateClosed
	}
	return dbBreaker.State()
}

// RedisState 获取 Redis 熔断器状态
func RedisState() gobreaker.State {
	if redisBreaker == nil {
		return gobreaker.StateClosed
	}
	return redisBreaker.State()
}

// IsCircuitOpen 任意熔断器是否打开
func IsCircuitOpen() bool {
	return (dbBreaker != nil && dbBreaker.State() == gobreaker.StateOpen) ||
		(redisBreaker != nil && redisBreaker.State() == gobreaker.StateOpen)
}

// RedisBreakerEnabled 是否启用 Redis 熔断器
func RedisBreakerEnabled() bool {
	return redisBreaker != nil
}
