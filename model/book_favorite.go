package model

import "time"

// BookFavorite 图书收藏表模型
type BookFavorite struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_user_id;uniqueIndex:uk_user_book;comment:用户ID"`
	BookID    uint      `json:"book_id" gorm:"not null;index:idx_book_id;uniqueIndex:uk_user_book;comment:图书ID"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_created_at;comment:收藏时间"`
}

// TableName 指定表名
func (BookFavorite) TableName() string {
	return "book_favorites"
}

// BookFavoriteWithBook 收藏记录（包含图书信息）
type BookFavoriteWithBook struct {
	BookFavorite
	Book Book `json:"book" gorm:"foreignKey:BookID"`
}

// FavoriteStatus 收藏状态响应
type FavoriteStatus struct {
	IsFavorited   bool `json:"is_favorited"`   // 是否已收藏
	FavoriteCount int  `json:"favorite_count"` // 收藏总数
}

// BatchFavoriteStatus 批量收藏状态响应
type BatchFavoriteStatus struct {
	BookID        uint `json:"book_id"`
	IsFavorited   bool `json:"is_favorited"`
	FavoriteCount int  `json:"favorite_count"`
}
