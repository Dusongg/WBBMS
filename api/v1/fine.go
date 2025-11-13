package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FineApi struct{}

var fineService = &service.FineService{}

// GetMyFines 获取我的罚款列表
func (f *FineApi) GetMyFines(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 查找读者ID
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 获取罚款列表
	fines, err := fineService.GetReaderUnpaidFines(reader.ID)
	if err != nil {
		global.GVA_LOG.Error("获取罚款列表失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取罚款列表失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"list":         fines,
		"total_unpaid": reader.UnpaidFine,
	}))
}

// GetFineList 获取罚款列表（管理员）
func (f *FineApi) GetFineList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var fines []model.FineRecord
	var total int64
	db := global.GVA_DB.Model(&model.FineRecord{}).
		Preload("Reader").
		Preload("Reader.User").
		Preload("BorrowRecord").
		Preload("BorrowRecord.Book")

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + pageInfo.Keyword + "%"
		db = db.Joins("JOIN readers ON fine_records.reader_id = readers.id").
			Where("readers.reader_no LIKE ?", keyword)
	}

	// 状态筛选
	status := c.Query("status")
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Order("fine_records.created_at DESC").
		Limit(pageInfo.PageSize).
		Offset(offset).
		Find(&fines).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     fines,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// PayFine 支付罚款
func (f *FineApi) PayFine(c *gin.Context) {
	var req struct {
		FineID     uint    `json:"fine_id" binding:"required"`
		PaidAmount float64 `json:"paid_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 获取操作员ID
	operatorID, _ := c.Get("user_id")
	var opID uint
	if operatorID != nil {
		opID = operatorID.(uint)
	}

	// 支付罚款
	if err := fineService.PayFine(req.FineID, req.PaidAmount, opID); err != nil {
		global.GVA_LOG.Error("支付罚款失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("支付成功"))
}

// WaiveFine 豁免罚款（管理员）
func (f *FineApi) WaiveFine(c *gin.Context) {
	fineIDStr := c.Param("id")
	fineID, err := strconv.ParseUint(fineIDStr, 10, 32)
	if err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&req)

	// 获取操作员ID
	operatorID, _ := c.Get("user_id")
	var opID uint
	if operatorID != nil {
		opID = operatorID.(uint)
	}

	// 豁免罚款
	if err := fineService.WaiveFine(uint(fineID), opID, req.Remark); err != nil {
		global.GVA_LOG.Error("豁免罚款失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("豁免成功"))
}

