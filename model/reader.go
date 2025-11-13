package model

import (
	"gorm.io/gorm"
)

// ReaderStatus 读者状态
type ReaderStatus string

const (
	ReaderStatusPending  ReaderStatus = "pending"  // 待审核
	ReaderStatusActive   ReaderStatus = "active"   // 正常
	ReaderStatusInactive ReaderStatus = "inactive" // 停用
	ReaderStatusRejected ReaderStatus = "rejected" // 已拒绝
)

// Reader 读者信息表
type Reader struct {
	gorm.Model
	UserID          uint         `json:"user_id" gorm:"uniqueIndex;not null;comment:用户ID"`
	User            User         `json:"user" gorm:"foreignKey:UserID"`
	ReaderNo        string       `json:"reader_no" gorm:"uniqueIndex;not null;comment:读者编号"`
	IDCard          string       `json:"id_card" gorm:"uniqueIndex;comment:身份证号"`
	Address         string       `json:"address" gorm:"comment:地址"`
	Status          ReaderStatus `json:"status" gorm:"type:enum('pending','active','inactive','rejected');default:'pending';comment:状态"`
	MaxBorrow       int          `json:"max_borrow" gorm:"default:5;comment:最大借阅数量"`
	BorrowDays      int          `json:"borrow_days" gorm:"default:30;comment:借阅天数"`
	MaxRenew        int          `json:"max_renew" gorm:"default:2;comment:最大续借次数"`
	RenewDays       int          `json:"renew_days" gorm:"default:15;comment:每次续借延长天数"`
	MaxReservations int          `json:"max_reservations" gorm:"default:3;comment:最大预约数量"`
	TotalFine       float64      `json:"total_fine" gorm:"default:0;comment:累计罚款"`
	UnpaidFine      float64      `json:"unpaid_fine" gorm:"default:0;comment:未支付罚款"`
	IsBlacklisted   bool         `json:"is_blacklisted" gorm:"default:false;comment:是否在黑名单"`
	Remark          string       `json:"remark" gorm:"type:text;comment:备注"`
}

func (Reader) TableName() string {
	return "readers"
}
