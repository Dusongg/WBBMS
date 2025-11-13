package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BlacklistApi struct{}

var blacklistService = &service.BlacklistService{}

// GetBlacklistList 获取黑名单列表（管理员）
func (b *BlacklistApi) GetBlacklistList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var blacklists []model.Blacklist
	var total int64
	db := global.GVA_DB.Model(&model.Blacklist{}).
		Preload("Reader").
		Preload("Reader.User")

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + pageInfo.Keyword + "%"
		db = db.Joins("JOIN readers ON blacklists.reader_id = readers.id").
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
	if err := db.Order("blacklists.created_at DESC").
		Limit(pageInfo.PageSize).
		Offset(offset).
		Find(&blacklists).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     blacklists,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// AddToBlacklist 添加到黑名单（管理员）
func (b *BlacklistApi) AddToBlacklist(c *gin.Context) {
	var req struct {
		ReaderID    uint   `json:"reader_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
		EndDate     string `json:"end_date"` // 可选，格式：2024-01-01，为空表示永久
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

	// 解析结束日期
	var endDate *time.Time
	if req.EndDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(200, response.FailWithMessage("日期格式错误"))
			return
		}
		endDate = &parsedDate
	}

	// 添加到黑名单
	reason := model.BlacklistReason(req.Reason)
	if err := blacklistService.AddToBlacklist(req.ReaderID, reason, req.Description, endDate, opID); err != nil {
		global.GVA_LOG.Error("添加黑名单失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("添加黑名单成功"))
}

// RemoveFromBlacklist 解除黑名单（管理员）
func (b *BlacklistApi) RemoveFromBlacklist(c *gin.Context) {
	blacklistIDStr := c.Param("id")
	blacklistID, err := strconv.ParseUint(blacklistIDStr, 10, 32)
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

	// 解除黑名单
	if err := blacklistService.RemoveFromBlacklist(uint(blacklistID), opID, req.Remark); err != nil {
		global.GVA_LOG.Error("解除黑名单失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("解除黑名单成功"))
}

// GetMyBlacklistStatus 获取我的黑名单状态
func (b *BlacklistApi) GetMyBlacklistStatus(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 查找读者信息
	var reader model.Reader
	if err := global.GVA_DB.Where("user_id = ?", userID).First(&reader).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者信息不存在"))
		return
	}

	// 获取黑名单记录
	blacklists, err := blacklistService.GetReaderBlacklists(reader.ID)
	if err != nil {
		global.GVA_LOG.Error("获取黑名单记录失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取黑名单记录失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"is_blacklisted": reader.IsBlacklisted,
		"blacklists":     blacklists,
	}))
}

