package model

import (
	"time"

	"gorm.io/gorm"
)

// BookCategory 图书分类中间表
type BookCategory struct {
	BookID     uint     `gorm:"primaryKey;comment:图书ID"`
	CategoryID uint     `gorm:"primaryKey;comment:分类ID"`
	Book       Book     `gorm:"foreignKey:BookID"`
	Category   Category `gorm:"foreignKey:CategoryID"`
}

func (BookCategory) TableName() string {
	return "book_categories"
}

type Book struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Title          string         `json:"title" gorm:"not null;comment:书名"`
	Author         string         `json:"author" gorm:"not null;comment:作者"`
	Publisher      string         `json:"publisher" gorm:"comment:出版社"`
	PublishDate    string         `json:"publish_date" gorm:"comment:出版日期"`
	ISBN           string         `json:"isbn" gorm:"uniqueIndex;comment:ISBN"`
	Price          float64        `json:"price" gorm:"comment:价格"`
	Description    string         `json:"description" gorm:"type:text;comment:描述"`
	Category       string         `json:"category" gorm:"comment:分类（兼容旧数据）"`
	Categories     []Category     `json:"categories" gorm:"many2many:book_categories;comment:分类列表"`
	CoverImage     string         `json:"cover_image" gorm:"comment:封面图片URL"`
	TotalStock     int            `json:"total_stock" gorm:"default:0;comment:总库存"`
	AvailableStock int            `json:"available_stock" gorm:"default:0;comment:可借库存"`
	LikeCount      int            `json:"like_count" gorm:"default:0;index;comment:点赞总数"`
	FavoriteCount  int            `json:"favorite_count" gorm:"default:0;index;comment:收藏总数"`
}

func (Book) TableName() string {
	return "books"
}
