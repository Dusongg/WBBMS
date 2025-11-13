package model

import (
	"time"

	"gorm.io/gorm"
)

// ReservationStatus 预约状态
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"   // 等待中
	ReservationStatusAvailable ReservationStatus = "available" // 可取书
	ReservationStatusFulfilled ReservationStatus = "fulfilled" // 已完成
	ReservationStatusCancelled ReservationStatus = "cancelled" // 已取消
	ReservationStatusExpired   ReservationStatus = "expired"   // 已过期
)

// Reservation 图书预约表
type Reservation struct {
	gorm.Model
	ReaderID       uint              `json:"reader_id" gorm:"not null;comment:读者ID;index:idx_reader_book"`
	Reader         Reader            `json:"reader" gorm:"foreignKey:ReaderID"`
	BookID         uint              `json:"book_id" gorm:"not null;comment:图书ID;index:idx_reader_book"`
	Book           Book              `json:"book" gorm:"foreignKey:BookID"`
	Status         ReservationStatus `json:"status" gorm:"type:enum('pending','available','fulfilled','cancelled','expired');default:'pending';comment:状态"`
	ReserveDate    time.Time         `json:"reserve_date" gorm:"comment:预约日期"`
	AvailableDate  *time.Time        `json:"available_date" gorm:"comment:可取书日期"`
	PickupDeadline *time.Time        `json:"pickup_deadline" gorm:"comment:取书截止日期"`
	FulfilledDate  *time.Time        `json:"fulfilled_date" gorm:"comment:完成日期（取书或取消）"`
	QueuePosition  int               `json:"queue_position" gorm:"comment:队列位置"`
	Remark         string            `json:"remark" gorm:"type:text;comment:备注"`
}

func (Reservation) TableName() string {
	return "reservations"
}
