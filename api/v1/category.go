package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CategoryApi struct{}

// GetCategoryList 获取分类列表
func (c *CategoryApi) GetCategoryList(ctx *gin.Context) {
	var categories []model.Category
	if err := global.GVA_DB.Order("sort ASC, name ASC").Find(&categories).Error; err != nil {
		global.GVA_LOG.Error("获取分类列表失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("获取分类列表失败: "+err.Error()))
		return
	}

	ctx.JSON(200, response.OkWithData(categories))
}

// CreateCategory 创建分类
func (c *CategoryApi) CreateCategory(ctx *gin.Context) {
	var category model.Category
	if err := ctx.ShouldBindJSON(&category); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 检查是否存在同名分类（包括软删除的）
	var existCategory model.Category
	if err := global.GVA_DB.Unscoped().Where("name = ?", category.Name).First(&existCategory).Error; err == nil {
		// 如果存在软删除的分类，恢复它
		if existCategory.DeletedAt.Valid {
			existCategory.Name = category.Name
			existCategory.Description = category.Description
			existCategory.Sort = category.Sort
			existCategory.DeletedAt = gorm.DeletedAt{} // 清除软删除标记
			if err := global.GVA_DB.Unscoped().Save(&existCategory).Error; err != nil {
				global.GVA_LOG.Error("恢复分类失败", zap.Error(err))
				ctx.JSON(200, response.FailWithMessage("恢复分类失败: "+err.Error()))
				return
			}
			ctx.JSON(200, response.OkWithMessage("分类已恢复"))
			return
		}
		// 如果存在未删除的分类，返回错误
		global.GVA_LOG.Error("分类已存在", zap.String("name", category.Name))
		ctx.JSON(200, response.FailWithMessage("分类名称已存在"))
		return
	}

	// 创建新分类
	if err := global.GVA_DB.Create(&category).Error; err != nil {
		global.GVA_LOG.Error("创建失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("创建失败: "+err.Error()))
		return
	}

	ctx.JSON(200, response.OkWithMessage("创建成功"))
}

// UpdateCategory 更新分类
func (c *CategoryApi) UpdateCategory(ctx *gin.Context) {
	var req struct {
		ID          uint   `json:"id" binding:"required"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Sort        int    `json:"sort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var category model.Category
	if err := global.GVA_DB.First(&category, req.ID).Error; err != nil {
		global.GVA_LOG.Error("分类不存在", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("分类不存在"))
		return
	}

	updateData := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"sort":        req.Sort,
	}

	if err := global.GVA_DB.Model(&category).Updates(updateData).Error; err != nil {
		global.GVA_LOG.Error("更新失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("更新失败: "+err.Error()))
		return
	}

	ctx.JSON(200, response.OkWithMessage("更新成功"))
}

// DeleteCategory 删除分类
func (c *CategoryApi) DeleteCategory(ctx *gin.Context) {
	var req request.GetById
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 检查是否有图书使用此分类
	var count int64
	global.GVA_DB.Model(&model.BookCategory{}).Where("category_id = ?", req.ID).Count(&count)
	if count > 0 {
		ctx.JSON(200, response.FailWithMessage("该分类下还有图书，无法删除"))
		return
	}

	// 使用硬删除，真正删除记录，避免软删除后无法创建同名分类
	if err := global.GVA_DB.Unscoped().Delete(&model.Category{}, req.ID).Error; err != nil {
		global.GVA_LOG.Error("删除失败", zap.Error(err))
		ctx.JSON(200, response.FailWithMessage("删除失败: "+err.Error()))
		return
	}

	ctx.JSON(200, response.OkWithMessage("删除成功"))
}
