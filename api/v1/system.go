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

type SystemApi struct{}

// GetUserList 获取用户列表
func (s *SystemApi) GetUserList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)

	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}

	var users []model.User
	var total int64
	db := global.GVA_DB.Model(&model.User{})

	// 搜索功能
	if pageInfo.Keyword != "" {
		keyword := "%" + strings.TrimSpace(pageInfo.Keyword) + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?", keyword, keyword, keyword)
	}

	if err := db.Count(&total).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	if err := db.Order("created_at DESC").Limit(pageInfo.PageSize).Offset(offset).Find(&users).Error; err != nil {
		global.GVA_LOG.Error("获取数据失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取数据失败"))
		return
	}

	c.JSON(200, response.OkWithDetailed(response.PageResult{
		List:     users,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功"))
}

// CreateUser 创建用户
func (s *SystemApi) CreateUser(c *gin.Context) {
	var req struct {
		Username string         `json:"username" binding:"required"`
		Password string         `json:"password" binding:"required"`
		Email    string         `json:"email"`
		Phone    string         `json:"phone"`
		Role     model.UserRole `json:"role" binding:"required"`
		RealName string         `json:"real_name"`
		Status   string         `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 检查用户名是否已存在
	var existUser model.User
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&existUser).Error; err == nil {
		c.JSON(200, response.FailWithMessage("用户名已存在"))
		return
	}

	user := model.User{
		Username: req.Username,
		Password: req.Password, // 简化处理，实际应加密
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     req.Role,
		RealName: req.RealName,
		Status:   "active",
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	if err := global.GVA_DB.Create(&user).Error; err != nil {
		global.GVA_LOG.Error("创建用户失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("创建失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("创建成功"))
}

// UpdateUser 更新用户
func (s *SystemApi) UpdateUser(c *gin.Context) {
	var req struct {
		ID       uint           `json:"id" binding:"required"`
		Email    string         `json:"email"`
		Phone    string         `json:"phone"`
		Role     model.UserRole `json:"role"`
		RealName string         `json:"real_name"`
		Status   string         `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	var user model.User
	if err := global.GVA_DB.First(&user, req.ID).Error; err != nil {
		c.JSON(200, response.FailWithMessage("用户不存在"))
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := global.GVA_DB.Save(&user).Error; err != nil {
		global.GVA_LOG.Error("更新用户失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("更新失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("更新成功"))
}

// DeleteUser 删除用户
func (s *SystemApi) DeleteUser(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	if err := global.GVA_DB.Delete(&model.User{}, req.ID).Error; err != nil {
		global.GVA_LOG.Error("删除用户失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("删除失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("删除成功"))
}

