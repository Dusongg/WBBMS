package model

import (
	"time"

	"gorm.io/gorm"
)

// BorrowStatus 借阅状态
type BorrowStatus string

const (
	BorrowStatusPending  BorrowStatus = "pending"  // 待批准
	BorrowStatusBorrowed BorrowStatus = "borrowed" // 已借出
	BorrowStatusReturned BorrowStatus = "returned" // 已归还
	BorrowStatusOverdue  BorrowStatus = "overdue"  // 逾期
	BorrowStatusRenewed  BorrowStatus = "renewed"  // 已续借
	BorrowStatusReserved BorrowStatus = "reserved" // 已预约
	BorrowStatusRejected BorrowStatus = "rejected" // 已拒绝
)

// BorrowRecord 借阅记录表
type BorrowRecord struct {
	gorm.Model
	ReaderID      uint         `json:"reader_id" gorm:"not null;comment:读者ID;index"`
	Reader        Reader       `json:"reader" gorm:"foreignKey:ReaderID"`
	BookID        uint         `json:"book_id" gorm:"not null;comment:图书ID;index"`
	Book          Book         `json:"book" gorm:"foreignKey:BookID"`
	BorrowDate    time.Time    `json:"borrow_date" gorm:"comment:借阅日期"`
	DueDate       time.Time    `json:"due_date" gorm:"comment:应还日期;index"`
	ReturnDate    *time.Time   `json:"return_date" gorm:"comment:实际归还日期"`
	Status        BorrowStatus `json:"status" gorm:"type:enum('pending','borrowed','returned','overdue','renewed','reserved','rejected');default:'pending';comment:状态;index"`
	RenewCount    int          `json:"renew_count" gorm:"default:0;comment:续借次数"`
	MaxRenewCount int          `json:"max_renew_count" gorm:"default:2;comment:最大续借次数"`
	FineAmount    float64      `json:"fine_amount" gorm:"default:0;comment:罚款金额"`
	OverdueDays   int          `json:"overdue_days" gorm:"default:0;comment:逾期天数"`
	ReservationID *uint        `json:"reservation_id" gorm:"comment:预约ID（如果是通过预约借书）"`
	OperatorID    uint         `json:"operator_id" gorm:"comment:操作员ID"`
	Remark        string       `json:"remark" gorm:"type:text;comment:备注"`
}

func (BorrowRecord) TableName() string {
	return "borrow_records"
}
