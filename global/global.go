package global

import (
	"bookadmin/config"
	"bookadmin/model"
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GVA_DB     *gorm.DB
	GVA_LOG    *zap.Logger
	GVA_REDIS  *redis.Client
	GVA_CONFIG *config.Config
	BookModel  model.Book
)

// DB 返回带 context 的 DB 实例，用于分布式追踪链路关联。
// 当 ctx 为 nil 时返回原始 GVA_DB（兼容非请求场景如 migrate、cron）。
func DB(ctx context.Context) *gorm.DB {
	if GVA_DB == nil {
		return nil
	}
	if ctx == nil {
		return GVA_DB
	}
	return GVA_DB.WithContext(ctx)
}
