package initialize

import (
	"bookadmin/global"
	"bookadmin/model"

	"go.uber.org/zap"
)

// InitData 初始化默认数据
func InitData() {
	// 创建默认管理员账户
	var adminUser model.User
	if err := global.GVA_DB.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		adminUser = model.User{
			Username: "admin",
			Password: "admin123", // 简化处理，实际应加密
			Email:    "admin@bookadmin.com",
			Role:     model.RoleAdmin,
			Status:   "active",
			RealName: "系统管理员",
		}
		if err := global.GVA_DB.Create(&adminUser).Error; err != nil {
			global.GVA_LOG.Error("创建默认管理员失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("默认管理员创建成功: admin/admin123")
		}
	}

	// 创建默认图书管理员
	var librarianUser model.User
	if err := global.GVA_DB.Where("username = ?", "librarian").First(&librarianUser).Error; err != nil {
		librarianUser = model.User{
			Username: "librarian",
			Password: "librarian123",
			Email:    "librarian@bookadmin.com",
			Role:     model.RoleLibrarian,
			Status:   "active",
			RealName: "图书管理员",
		}
		if err := global.GVA_DB.Create(&librarianUser).Error; err != nil {
			global.GVA_LOG.Error("创建默认图书管理员失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("默认图书管理员创建成功: librarian/librarian123")
		}
	}

	// 创建默认分类
	defaultCategories := []model.Category{
		{Name: "国内", Description: "国内图书", Sort: 1},
		{Name: "国外", Description: "国外图书", Sort: 2},
		{Name: "儿童", Description: "儿童图书", Sort: 3},
		{Name: "其他", Description: "其他分类", Sort: 99},
	}

	for _, category := range defaultCategories {
		var existCategory model.Category
		if err := global.GVA_DB.Where("name = ?", category.Name).First(&existCategory).Error; err != nil {
			if err := global.GVA_DB.Create(&category).Error; err != nil {
				global.GVA_LOG.Error("创建默认分类失败", zap.String("name", category.Name), zap.Error(err))
			} else {
				global.GVA_LOG.Info("创建默认分类成功", zap.String("name", category.Name))
			}
		}
	}
}
