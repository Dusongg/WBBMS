package model

import "time"

// BookLike 图书点赞表模型
type BookLike struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_user_id;uniqueIndex:uk_user_book;comment:用户ID"`
	BookID    uint      `json:"book_id" gorm:"not null;index:idx_book_id;uniqueIndex:uk_user_book;comment:图书ID"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_created_at;comment:点赞时间"`
}

// TableName 指定表名
func (BookLike) TableName() string {
	return "book_likes"
}

// BookLikeWithBook 点赞记录（包含图书信息）
type BookLikeWithBook struct {
	BookLike
	Book Book `json:"book" gorm:"foreignKey:BookID"`
}

// LikeStatus 点赞状态响应
type LikeStatus struct {
	IsLiked   bool `json:"is_liked"`   // 是否已点赞
	LikeCount int  `json:"like_count"` // 点赞总数
}

// BatchLikeStatus 批量点赞状态响应
type BatchLikeStatus struct {
	BookID    uint `json:"book_id"`
	IsLiked   bool `json:"is_liked"`
	LikeCount int  `json:"like_count"`
}
