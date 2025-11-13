package model

import (
	"time"

	"gorm.io/gorm"
)

// BlacklistStatus 黑名单状态
type BlacklistStatus string

const (
	BlacklistStatusActive   BlacklistStatus = "active"   // 生效中
	BlacklistStatusLifted   BlacklistStatus = "lifted"   // 已解除
	BlacklistStatusExpired  BlacklistStatus = "expired"  // 已过期
)

// BlacklistReason 拉黑原因
type BlacklistReason string

const (
	BlacklistReasonOverdue    BlacklistReason = "overdue"    // 逾期未还
	BlacklistReasonDamage     BlacklistReason = "damage"     // 损坏图书
	BlacklistReasonLost       BlacklistReason = "lost"       // 丢失图书
	BlacklistReasonUnpaidFine BlacklistReason = "unpaid_fine" // 未支付罚款
	BlacklistReasonOther      BlacklistReason = "other"      // 其他
)

// Blacklist 黑名单表
type Blacklist struct {
	gorm.Model
	ReaderID    uint            `json:"reader_id" gorm:"not null;comment:读者ID;index"`
	Reader      Reader          `json:"reader" gorm:"foreignKey:ReaderID"`
	Reason      BlacklistReason `json:"reason" gorm:"type:varchar(50);comment:拉黑原因"`
	Description string          `json:"description" gorm:"type:text;comment:详细描述"`
	Status      BlacklistStatus `json:"status" gorm:"type:enum('active','lifted','expired');default:'active';comment:状态"`
	StartDate   time.Time       `json:"start_date" gorm:"comment:开始日期"`
	EndDate     *time.Time      `json:"end_date" gorm:"comment:结束日期（永久为NULL）"`
	LiftedDate  *time.Time      `json:"lifted_date" gorm:"comment:解除日期"`
	OperatorID  uint            `json:"operator_id" gorm:"comment:操作员ID"`
	Remark      string          `json:"remark" gorm:"type:text;comment:备注"`
}

func (Blacklist) TableName() string {
	return "blacklists"
}

