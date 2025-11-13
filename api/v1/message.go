package v1

import (
	"bookadmin/global"
	"bookadmin/model/common/response"
	"bookadmin/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MessageApi struct{}

var messageService = &service.MessageService{}

// GetMessages 获取消息列表
func (m *MessageApi) GetMessages(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 查询消息列表
	messages, total, err := messageService.GetUserMessages(userID, page, pageSize)
	if err != nil {
		global.GVA_LOG.Error("获取消息列表失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取消息列表失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"list":     messages,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

// GetUnreadCount 获取未读消息数量
func (m *MessageApi) GetUnreadCount(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	count, err := messageService.GetUnreadCount(userID)
	if err != nil {
		global.GVA_LOG.Error("获取未读消息数量失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取未读消息数量失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"count": count,
	}))
}

// MarkAsRead 标记消息为已读
func (m *MessageApi) MarkAsRead(c *gin.Context) {
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	if err := messageService.MarkAsRead(uint(messageID), userID); err != nil {
		global.GVA_LOG.Error("标记消息已读失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("操作成功"))
}

// MarkAllAsRead 标记所有消息为已读
func (m *MessageApi) MarkAllAsRead(c *gin.Context) {
	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	if err := messageService.MarkAllAsRead(userID); err != nil {
		global.GVA_LOG.Error("标记所有消息已读失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("操作失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("操作成功"))
}

// DeleteMessage 删除消息
func (m *MessageApi) DeleteMessage(c *gin.Context) {
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 获取当前用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}
	userID := userIDInterface.(uint)

	if err := messageService.DeleteMessage(uint(messageID), userID); err != nil {
		global.GVA_LOG.Error("删除消息失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage(err.Error()))
		return
	}

	c.JSON(200, response.OkWithMessage("删除成功"))
}
