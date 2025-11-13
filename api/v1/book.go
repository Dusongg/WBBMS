package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BookApi struct{}

// CreateBook 创建图书
func (b *BookApi) CreateBook(c *gin.Context) {
	var req struct {
		model.Book
		CategoryIDs []uint `json:"category_ids"` // 分类ID列表
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 开始事务
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建图书
	if err := tx.Create(&req.Book).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("创建失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("创建失败: "+err.Error()))
		return
	}

	// 关联分类
	if len(req.CategoryIDs) > 0 {
		var categories []model.Category
		if err := tx.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
			tx.Rollback()
			global.GVA_LOG.Error("查找分类失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("查找分类失败: "+err.Error()))
			return
		}
		if err := tx.Model(&req.Book).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			global.GVA_LOG.Error("关联分类失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("关联分类失败: "+err.Error()))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		global.GVA_LOG.Error("提交事务失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("创建失败: "+err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("创建成功"))
}

// DeleteBook 删除图书
func (b *BookApi) DeleteBook(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	if err := global.GVA_DB.Delete(&model.Book{}, req.ID).Error; err != nil {
		global.GVA_LOG.Error("删除失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("删除失败: "+err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("删除成功"))
}

// UpdateBook 更新图书
func (b *BookApi) UpdateBook(c *gin.Context) {
	var req struct {
		ID             uint    `json:"id"`
		Title          string  `json:"title"`
		Author         string  `json:"author"`
		Publisher      string  `json:"publisher"`
		PublishDate    string  `json:"publish_date"`
		ISBN           string  `json:"isbn"`
		Price          float64 `json:"price"`
		Description    string  `json:"description"`
		Category       string  `json:"category"`
		CategoryIDs    []uint  `json:"category_ids"` // 分类ID列表
		CoverImage     string  `json:"cover_image"`
		TotalStock     int     `json:"total_stock"`
		AvailableStock int     `json:"available_stock"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误: "+err.Error()))
		return
	}

	// 检查ID是否有效
	if req.ID == 0 {
		global.GVA_LOG.Error("图书ID无效", zap.Uint("id", req.ID), zap.Any("request", req))
		c.JSON(200, response.FailWithMessage("图书ID无效，请确保ID字段存在且大于0"))
		return
	}

	bookID := req.ID

	// 检查图书是否存在（排除软删除的记录）
	var existBook model.Book
	if err := global.GVA_DB.Where("id = ?", bookID).First(&existBook).Error; err != nil {
		global.GVA_LOG.Error("图书不存在", zap.Uint("id", bookID), zap.Error(err))
		c.JSON(200, response.FailWithMessage("图书不存在"))
		return
	}

	// 记录更新前的库存，用于检测库存变化
	oldAvailableStock := existBook.AvailableStock

	// 开始事务
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用map明确指定要更新的字段，包括cover_image（即使为空字符串也要更新）
	updateData := map[string]interface{}{
		"title":           req.Title,
		"author":          req.Author,
		"publisher":       req.Publisher,
		"publish_date":    req.PublishDate,
		"isbn":            req.ISBN,
		"price":           req.Price,
		"description":     req.Description,
		"category":        req.Category,
		"cover_image":     req.CoverImage,
		"total_stock":     req.TotalStock,
		"available_stock": req.AvailableStock,
	}

	if err := tx.Model(&existBook).Updates(updateData).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("更新失败", zap.Uint("id", bookID), zap.Error(err))
		c.JSON(200, response.FailWithMessage("更新失败: "+err.Error()))
		return
	}

	// 更新分类关联
	if req.CategoryIDs != nil {
		var categories []model.Category
		if len(req.CategoryIDs) > 0 {
			if err := tx.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
				tx.Rollback()
				global.GVA_LOG.Error("查找分类失败", zap.Error(err))
				c.JSON(200, response.FailWithMessage("查找分类失败: "+err.Error()))
				return
			}
		}
		if err := tx.Model(&existBook).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			global.GVA_LOG.Error("更新分类关联失败", zap.Error(err))
			c.JSON(200, response.FailWithMessage("更新分类关联失败: "+err.Error()))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		global.GVA_LOG.Error("提交事务失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("更新失败: "+err.Error()))
		return
	}

	// 检测库存变化：如果从0变为>0，检查并通知预约者
	// 重新查询更新后的实际库存值，确保准确性
	var updatedBook model.Book
	if err := global.GVA_DB.First(&updatedBook, bookID).Error; err == nil {
		// 如果原来库存是0，现在库存>0，说明有库存补充
		if oldAvailableStock == 0 && updatedBook.AvailableStock > 0 {
			// 异步处理预约通知，不影响更新响应
			go func() {
				reservationService := service.NewReservationService()
				if err := reservationService.CheckAndNotifyAvailableReservations(bookID); err != nil {
					global.GVA_LOG.Error("通知预约者失败", zap.Uint("book_id", bookID), zap.Error(err))
				} else {
					global.GVA_LOG.Info("库存补充，已通知预约者",
						zap.Uint("book_id", bookID),
						zap.Int("old_stock", oldAvailableStock),
						zap.Int("new_stock", updatedBook.AvailableStock))
				}
			}()
		}
	}

	c.JSON(200, response.OkWithMessage("更新成功"))
}

// GetBook 获取单个图书
func (b *BookApi) GetBook(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var book model.Book
	if err := global.GVA_DB.Preload("Categories").First(&book, req.ID).Error; err != nil {
		global.GVA_LOG.Error("获取失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取失败: "+err.Error()))
		return
	}

	c.JSON(200, response.OkWithData(book))
}

// GetBookList 获取图书列表，支持搜索和分页
func (b *BookApi) GetBookList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	// 设置默认值
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var books []model.Book
	var total int64
	db := global.GVA_DB.Model(&model.Book{})

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + strings.TrimSpace(pageInfo.Keyword) + "%"
		db = db.Where("title LIKE ? OR author LIKE ? OR publisher LIKE ? OR isbn LIKE ?",
			keyword, keyword, keyword, keyword)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败: "+err.Error()))
		return
	}

	// 分页查询，预加载分类
	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Preload("Categories").Order("created_at DESC").Limit(pageInfo.PageSize).Offset(offset).Find(&books).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败: "+err.Error()))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     books,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// DeleteBookById 通过ID删除图书（URL参数方式）
func (b *BookApi) DeleteBookById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		global.GVA_LOG.Error("参数错误", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	if err := global.GVA_DB.Delete(&model.Book{}, id).Error; err != nil {
		global.GVA_LOG.Error("删除失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("删除失败: "+err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("删除成功"))
}
