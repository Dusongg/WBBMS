package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/response"
	"time"

	"github.com/gin-gonic/gin"
)

type StatisticsApi struct{}

// GetStatistics 获取统计信息
func (s *StatisticsApi) GetStatistics(c *gin.Context) {
	var stats = make(map[string]interface{})

	// 图书总数
	var totalBooks int64
	global.GVA_DB.Model(&model.Book{}).Count(&totalBooks)
	stats["total_books"] = totalBooks

	// 可借图书数
	var availableBooks int64
	global.GVA_DB.Model(&model.Book{}).Where("available_stock > 0").Count(&availableBooks)
	stats["available_books"] = availableBooks

	// 读者总数
	var totalReaders int64
	global.GVA_DB.Model(&model.Reader{}).Where("status = ?", model.ReaderStatusActive).Count(&totalReaders)
	stats["total_readers"] = totalReaders

	// 借阅中数量
	var borrowingCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("status IN (?)", []model.BorrowStatus{model.BorrowStatusBorrowed, model.BorrowStatusOverdue}).
		Count(&borrowingCount)
	stats["borrowing_count"] = borrowingCount

	// 逾期数量
	var overdueCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("status = ? AND due_date < ?", model.BorrowStatusBorrowed, time.Now()).
		Count(&overdueCount)
	stats["overdue_count"] = overdueCount

	// 本月借阅数量
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthBorrowCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("borrow_date >= ?", startOfMonth).
		Count(&monthBorrowCount)
	stats["month_borrow_count"] = monthBorrowCount

	// 本月归还数量
	var monthReturnCount int64
	global.GVA_DB.Model(&model.BorrowRecord{}).
		Where("return_date >= ?", startOfMonth).
		Count(&monthReturnCount)
	stats["month_return_count"] = monthReturnCount

	c.JSON(200, response.OkWithData(stats))
}

// GetBorrowStatistics 获取借阅统计（按时间）
func (s *StatisticsApi) GetBorrowStatistics(c *gin.Context) {
	var req struct {
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}
	c.ShouldBindQuery(&req)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0) // 默认最近一个月
	endDate := now

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = t
		}
	}

	var records []model.BorrowRecord
	global.GVA_DB.Where("borrow_date >= ? AND borrow_date <= ?", startDate, endDate).
		Preload("Book").Preload("Reader.User").
		Find(&records)

	c.JSON(200, response.OkWithData(records))
}

// GetPopularBooks 获取热门图书（借阅次数最多）
func (s *StatisticsApi) GetPopularBooks(c *gin.Context) {
	var popularBooks []struct {
		BookID      uint   `json:"book_id"`
		Title       string `json:"title"`
		Author      string `json:"author"`
		BorrowCount int64  `json:"borrow_count"`
	}

	global.GVA_DB.Model(&model.BorrowRecord{}).
		Select("book_id, COUNT(*) as borrow_count").
		Group("book_id").
		Order("borrow_count DESC").
		Limit(10).
		Scan(&popularBooks)

	// 填充图书信息
	for i := range popularBooks {
		var book model.Book
		if err := global.GVA_DB.First(&book, popularBooks[i].BookID).Error; err == nil {
			popularBooks[i].Title = book.Title
			popularBooks[i].Author = book.Author
		}
	}

	c.JSON(200, response.OkWithData(popularBooks))
}
