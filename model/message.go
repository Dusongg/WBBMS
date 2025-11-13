package model

import (
	"time"

	"gorm.io/gorm"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeSystem      MessageType = "system"       // 系统消息
	MessageTypeReservation MessageType = "reservation"  // 预约通知
	MessageTypeBorrow      MessageType = "borrow"       // 借阅通知
	MessageTypeOverdue     MessageType = "overdue"      // 逾期提醒
	MessageTypeFine        MessageType = "fine"         // 罚款通知
)

// Message 站内消息
type Message struct {
	gorm.Model
	UserID     uint        `json:"user_id" gorm:"not null;index;comment:用户ID"`
	Type       MessageType `json:"type" gorm:"type:enum('system','reservation','borrow','overdue','fine');not null;comment:消息类型"`
	Title      string      `json:"title" gorm:"type:varchar(255);not null;comment:消息标题"`
	Content    string      `json:"content" gorm:"type:text;not null;comment:消息内容"`
	IsRead     bool        `json:"is_read" gorm:"default:false;comment:是否已读"`
	ReadAt     *time.Time  `json:"read_at" gorm:"comment:阅读时间"`
	RelatedID  *uint       `json:"related_id" gorm:"comment:关联ID（如预约ID、借阅ID等）"`
	RelatedType string     `json:"related_type" gorm:"type:varchar(50);comment:关联类型"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

