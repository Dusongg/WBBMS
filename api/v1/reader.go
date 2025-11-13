package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/request"
	"bookadmin/model/common/response"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReaderApi struct{}

// GetReaderList 获取读者列表（支持搜索和分页）
func (r *ReaderApi) GetReaderList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var readers []model.Reader
	var total int64
	db := global.GVA_DB.Model(&model.Reader{}).Preload("User")

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + strings.TrimSpace(pageInfo.Keyword) + "%"
		db = db.Joins("JOIN users ON readers.user_id = users.id").
			Where("users.username LIKE ? OR users.real_name LIKE ? OR readers.reader_no LIKE ? OR readers.id_card LIKE ?",
				keyword, keyword, keyword, keyword)
	}

	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Order("readers.created_at DESC").Limit(pageInfo.PageSize).Offset(offset).Find(&readers).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     readers,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// GetReader 获取单个读者信息
func (r *ReaderApi) GetReader(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var reader model.Reader
	if err := global.GVA_DB.Preload("User").First(&reader, req.ID).Error; err != nil {
		global.GVA_LOG.Error("获取读者信息失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取读者信息失败"))
		return
	}

	c.JSON(200, response.OkWithData(reader))
}

// UpdateReaderStatus 更新读者状态（审核）
func (r *ReaderApi) UpdateReaderStatus(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Status string `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 验证状态值
	validStatuses := []string{"pending", "active", "inactive", "rejected"}
	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(200, response.FailWithMessage("状态值无效"))
		return
	}

	var reader model.Reader
	if err := global.GVA_DB.First(&reader, req.ID).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者不存在"))
		return
	}

	reader.Status = model.ReaderStatus(req.Status)
	if req.Remark != "" {
		reader.Remark = req.Remark
	}

	if err := global.GVA_DB.Save(&reader).Error; err != nil {
		global.GVA_LOG.Error("更新读者状态失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("更新失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("更新成功"))
}

// UpdateReader 更新读者信息
func (r *ReaderApi) UpdateReader(c *gin.Context) {
	var req struct {
		ID        uint   `json:"id" binding:"required"`
		MaxBorrow   int    `json:"max_borrow"`
		BorrowDays  int    `json:"borrow_days"`
		Address     string `json:"address"`
		Remark      string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var reader model.Reader
	if err := global.GVA_DB.First(&reader, req.ID).Error; err != nil {
		c.JSON(200, response.FailWithMessage("读者不存在"))
		return
	}

	if req.MaxBorrow > 0 {
		reader.MaxBorrow = req.MaxBorrow
	}
	if req.BorrowDays > 0 {
		reader.BorrowDays = req.BorrowDays
	}
	if req.Address != "" {
		reader.Address = req.Address
	}
	if req.Remark != "" {
		reader.Remark = req.Remark
	}

	if err := global.GVA_DB.Save(&reader).Error; err != nil {
		global.GVA_LOG.Error("更新读者信息失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("更新失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("更新成功"))
}

