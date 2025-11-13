package model

import (
	"gorm.io/gorm"
)

// Category 分类表
type Category struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null;uniqueIndex;comment:分类名称"`
	Description string `json:"description" gorm:"type:text;comment:分类描述"`
	Sort        int    `json:"sort" gorm:"default:0;comment:排序"`
}

func (Category) TableName() string {
	return "categories"
}

