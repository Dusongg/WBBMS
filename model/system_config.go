package model

import (
	"gorm.io/gorm"
)

// SystemConfig 系统配置表
type SystemConfig struct {
	gorm.Model
	ConfigKey   string `json:"config_key" gorm:"uniqueIndex;not null;comment:配置键"`
	ConfigValue string `json:"config_value" gorm:"type:text;comment:配置值"`
	Description string `json:"description" gorm:"type:text;comment:配置描述"`
	ConfigType  string `json:"config_type" gorm:"type:varchar(50);comment:配置类型(int/float/string/bool)"` // int, float, string, bool
	IsSystem    bool   `json:"is_system" gorm:"default:false;comment:是否系统配置（系统配置不可删除）"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// 系统配置键常量
const (
	// 借阅规则
	ConfigMaxBorrowBooks = "max_borrow_books" // 最大借阅数量
	ConfigBorrowDays     = "borrow_days"      // 借阅天数
	ConfigMaxRenewTimes  = "max_renew_times"  // 最大续借次数
	ConfigRenewDays      = "renew_days"       // 续借延长天数

	// 预约规则
	ConfigMaxReservations       = "max_reservations"        // 最大预约数量
	ConfigReservationPickupDays = "reservation_pickup_days" // 预约取书有效期（天）

	// 罚款规则
	ConfigOverdueFinePerDay = "overdue_fine_per_day" // 逾期罚款（元/天）
	ConfigMaxFineRate       = "max_fine_rate"        // 罚款上限比例（图书价格的百分比）

	// 逾期规则
	ConfigOverdueReminderDays  = "overdue_reminder_days"  // 提前提醒天数
	ConfigOverdueBlockDays     = "overdue_block_days"     // 逾期多久禁止借书（天）
	ConfigOverdueBlacklistDays = "overdue_blacklist_days" // 逾期多久自动拉黑（天）
)
