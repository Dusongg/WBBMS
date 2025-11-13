package model

import (
	"time"

	"gorm.io/gorm"
)

// FineStatus 罚款状态
type FineStatus string

const (
	FineStatusUnpaid FineStatus = "unpaid" // 未支付
	FineStatusPaid   FineStatus = "paid"   // 已支付
	FineStatusWaived FineStatus = "waived" // 已豁免
)

// FineRecord 罚款记录表
type FineRecord struct {
	gorm.Model
	ReaderID       uint       `json:"reader_id" gorm:"not null;comment:读者ID;index"`
	Reader         Reader     `json:"reader" gorm:"foreignKey:ReaderID"`
	BorrowRecordID uint       `json:"borrow_record_id" gorm:"comment:借阅记录ID"`
	BorrowRecord   *BorrowRecord `json:"borrow_record,omitempty" gorm:"foreignKey:BorrowRecordID"`
	FineType       string     `json:"fine_type" gorm:"type:varchar(50);comment:罚款类型(overdue/damage/lost)"` // overdue逾期 damage损坏 lost丢失
	Amount         float64    `json:"amount" gorm:"not null;comment:罚款金额"`
	PaidAmount     float64    `json:"paid_amount" gorm:"default:0;comment:已支付金额"`
	Status         FineStatus `json:"status" gorm:"type:enum('unpaid','paid','waived');default:'unpaid';comment:状态"`
	OverdueDays    int        `json:"overdue_days" gorm:"comment:逾期天数"`
	FineDate       time.Time  `json:"fine_date" gorm:"comment:罚款日期"`
	PaidDate       *time.Time `json:"paid_date" gorm:"comment:支付日期"`
	OperatorID     uint       `json:"operator_id" gorm:"comment:操作员ID"`
	Remark         string     `json:"remark" gorm:"type:text;comment:备注"`
}

func (FineRecord) TableName() string {
	return "fine_records"
}

