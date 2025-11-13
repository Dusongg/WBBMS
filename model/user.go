package model

import (
	"gorm.io/gorm"
)

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin      UserRole = "admin"      // 系统管理员
	RoleLibrarian  UserRole = "librarian"  // 图书管理员
	RoleReader     UserRole = "reader"     // 普通读者
)

// User 系统用户表
type User struct {
	gorm.Model
	Username string   `json:"username" gorm:"uniqueIndex;not null;comment:用户名"`
	Password string   `json:"-" gorm:"not null;comment:密码"`
	Email    string   `json:"email" gorm:"uniqueIndex;comment:邮箱"`
	Phone    string   `json:"phone" gorm:"comment:手机号"`
	Role     UserRole `json:"role" gorm:"type:enum('admin','librarian','reader');default:'reader';comment:角色"`
	Status   string   `json:"status" gorm:"default:'active';comment:状态:active,inactive"`
	RealName string   `json:"real_name" gorm:"comment:真实姓名"`
}

func (User) TableName() string {
	return "users"
}

