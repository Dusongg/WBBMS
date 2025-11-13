package v1

import (
	"bookadmin/model"
	"bookadmin/service"
	"bookadmin/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LikeAPI 点赞API
type LikeAPI struct {
	service *service.LikeService
}

// NewLikeAPI 创建点赞API实例
func NewLikeAPI() *LikeAPI {
	return &LikeAPI{
		service: service.NewLikeService(),
	}
}

// ToggleLike 切换点赞状态
// @Summary 切换点赞状态
// @Tags Like
// @Accept json
// @Produce json
// @Param bookId path int true "图书ID"
// @Success 200 {object} model.LikeStatus
// @Router /api/v1/like/toggle/{bookId} [post]
func (api *LikeAPI) ToggleLike(c *gin.Context) {
	// 获取当前用户ID (注意：JWT中间件设置的是 user_id)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	// 获取图书ID
	bookIDStr := c.Param("bookId")
	bookID, err := strconv.ParseUint(bookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	// 执行切换操作
	status, err := api.service.ToggleLike(c.Request.Context(), userID.(uint), uint(bookID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "操作成功",
		"data": status,
	})
}

// GetLikeStatus 查询点赞状态
// @Summary 查询点赞状态
// @Tags Like
// @Produce json
// @Param bookId path int true "图书ID"
// @Success 200 {object} model.LikeStatus
// @Router /api/v1/like/status/{bookId} [get]
func (api *LikeAPI) GetLikeStatus(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	// 获取图书ID
	bookIDStr := c.Param("bookId")
	bookID, err := strconv.ParseUint(bookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	// 查询状态
	status, err := api.service.GetLikeStatus(c.Request.Context(), userID.(uint), uint(bookID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": status,
	})
}

// BatchGetLikeStatus 批量查询点赞状态
// @Summary 批量查询点赞状态
// @Tags Like
// @Accept json
// @Produce json
// @Param bookIds query string true "图书ID列表，逗号分隔"
// @Success 200 {array} model.BatchLikeStatus
// @Router /api/v1/like/batch-status [get]
func (api *LikeAPI) BatchGetLikeStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	// 获取图书ID列表
	bookIDsStr := c.Query("bookIds")
	if bookIDsStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []model.BatchLikeStatus{}})
		return
	}

	bookIDs, err := utils.ParseUintSlice(bookIDsStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	// 批量查询
	statusList, err := api.service.BatchGetLikeStatus(c.Request.Context(), userID.(uint), bookIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": statusList,
	})
}

// GetUserLikeList 获取用户点赞列表
// @Summary 获取用户点赞列表
// @Tags Like
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} model.PageResult
// @Router /api/v1/like/list [get]
func (api *LikeAPI) GetUserLikeList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 查询列表
	list, total, err := api.service.GetUserLikeList(c.Request.Context(), userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
