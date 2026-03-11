package global

import (
	"bookadmin/config"
	"bookadmin/model"

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
