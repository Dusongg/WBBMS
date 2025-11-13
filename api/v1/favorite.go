package v1

import (
	"bookadmin/model"
	"bookadmin/service"
	"bookadmin/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FavoriteAPI 收藏API
type FavoriteAPI struct {
	service *service.FavoriteService
}

// NewFavoriteAPI 创建收藏API实例
func NewFavoriteAPI() *FavoriteAPI {
	return &FavoriteAPI{
		service: service.NewFavoriteService(),
	}
}

// ToggleFavorite 切换收藏状态
func (api *FavoriteAPI) ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	bookIDStr := c.Param("bookId")
	bookID, err := strconv.ParseUint(bookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	status, err := api.service.ToggleFavorite(c.Request.Context(), userID.(uint), uint(bookID))
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

// GetFavoriteStatus 查询收藏状态
func (api *FavoriteAPI) GetFavoriteStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	bookIDStr := c.Param("bookId")
	bookID, err := strconv.ParseUint(bookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	status, err := api.service.GetFavoriteStatus(c.Request.Context(), userID.(uint), uint(bookID))
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

// BatchGetFavoriteStatus 批量查询收藏状态
func (api *FavoriteAPI) BatchGetFavoriteStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	bookIDsStr := c.Query("bookIds")
	if bookIDsStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []model.BatchFavoriteStatus{}})
		return
	}

	bookIDs, err := utils.ParseUintSlice(bookIDsStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图书ID格式错误"})
		return
	}

	statusList, err := api.service.BatchGetFavoriteStatus(c.Request.Context(), userID.(uint), bookIDs)
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

// GetUserFavoriteList 获取用户收藏列表
func (api *FavoriteAPI) GetUserFavoriteList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := api.service.GetUserFavoriteList(c.Request.Context(), userID.(uint), page, pageSize)
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
