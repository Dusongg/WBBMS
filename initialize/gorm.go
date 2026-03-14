package initialize

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/observability"
	"bookadmin/service"
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
)

func Gorm() *gorm.DB {
	m := global.GVA_DB
	if m == nil {
		return nil
	}

	// 自动迁移
	err := m.AutoMigrate(
		&model.Book{},
		&model.Category{},
		&model.BookCategory{},
		&model.User{},
		&model.Reader{},
		&model.BorrowRecord{},
		&model.BookLike{},     // 点赞表
		&model.BookFavorite{}, // 收藏表
		&model.Reservation{},  // 预约表
		&model.FineRecord{},   // 罚款记录表
		&model.Blacklist{},    // 黑名单表
		&model.SystemConfig{}, // 系统配置表
		&model.Message{},      // 消息表
	)
	if err != nil {
		global.GVA_LOG.Error("自动迁移失败", zap.Error(err))
	}

	// 初始化系统配置
	InitSystemConfigs(m)

	return m
}

// InitSystemConfigs 初始化系统配置
func InitSystemConfigs(db *gorm.DB) {
	// 检查是否已经初始化
	var count int64
	db.Model(&model.SystemConfig{}).Count(&count)
	if count > 0 {
		// 已经有配置数据，跳过初始化
		return
	}

	// 初始化默认配置
	configs := []model.SystemConfig{
		{ConfigKey: model.ConfigMaxBorrowBooks, ConfigValue: "5", Description: "每个用户最多可借图书数量", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigBorrowDays, ConfigValue: "30", Description: "图书借阅期限（天）", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigMaxRenewTimes, ConfigValue: "2", Description: "最大续借次数", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigRenewDays, ConfigValue: "15", Description: "每次续借延长天数", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigMaxReservations, ConfigValue: "3", Description: "每个用户最多预约数量", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigReservationPickupDays, ConfigValue: "3", Description: "预约取书有效期（天）", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigOverdueFinePerDay, ConfigValue: "0.5", Description: "逾期罚款（元/天）", ConfigType: "float", IsSystem: true},
		{ConfigKey: model.ConfigMaxFineRate, ConfigValue: "0.5", Description: "罚款上限比例（图书价格的百分比）", ConfigType: "float", IsSystem: true},
		{ConfigKey: model.ConfigOverdueReminderDays, ConfigValue: "3", Description: "到期前提前提醒天数", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigOverdueBlockDays, ConfigValue: "7", Description: "逾期多久后禁止借书（天）", ConfigType: "int", IsSystem: true},
		{ConfigKey: model.ConfigOverdueBlacklistDays, ConfigValue: "30", Description: "逾期多久后自动拉黑（天）", ConfigType: "int", IsSystem: true},
	}

	if err := db.Create(&configs).Error; err != nil {
		global.GVA_LOG.Error("初始化系统配置失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info("系统配置初始化成功", zap.Int("count", len(configs)))
	}
}

// InitConfigCache 初始化配置缓存
func InitConfigCache() {
	// 刷新配置缓存
	if err := service.GlobalConfigService.RefreshCache(context.Background()); err != nil {
		global.GVA_LOG.Error("配置缓存初始化失败", zap.Error(err))
	} else {
		global.GVA_LOG.Info("配置缓存初始化成功")
	}
}

func GormMysql() *gorm.DB {
	m := global.GVA_DB
	if m != nil {
		return m
	}

	cfg := global.GVA_CONFIG
	if cfg == nil {
		global.GVA_LOG.Error("配置未初始化，无法连接MySQL")
		return nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Path,
		cfg.MySQL.DBName,
		cfg.MySQL.Config,
	)

	logMode := logger.Silent
	if cfg.MySQL.LogMode {
		logMode = logger.Info
	}

	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         191,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}

	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logMode)}
	if observability.TracerEnabled() {
		gormCfg.Plugins = map[string]gorm.Plugin{
			"otelgorm": otelgorm.NewPlugin(),
		}
	}
	appCfg := global.GVA_CONFIG
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormCfg); err != nil {
		global.GVA_LOG.Error("MySQL启动异常", zap.Error(err))
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(appCfg.MySQL.MaxIdleConns)
		sqlDB.SetMaxOpenConns(appCfg.MySQL.MaxOpenConns)
		global.GVA_DB = db
		return db
	}
}
