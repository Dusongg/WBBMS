package v1

import (
	"bookadmin/model"
	"bookadmin/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RankingAPI 榜单API
type RankingAPI struct {
	service *service.RankingService
}

// NewRankingAPI 创建榜单API实例
func NewRankingAPI() *RankingAPI {
	return &RankingAPI{
		service: service.NewRankingService(),
	}
}

// GetRanking 获取榜单
func (api *RankingAPI) GetRanking(c *gin.Context) {
	var req model.RankingRequest

	// 绑定参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 默认返回100条
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 100
	}

	// 查询榜单
	ranking, err := api.service.GetRanking(c.Request.Context(), req.Type, req.Period, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": ranking,
	})
}

// GetLikeWeekRanking 获取点赞周榜
func (api *RankingAPI) GetLikeWeekRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	ranking, err := api.service.GetRanking(
		c.Request.Context(),
		model.RankingTypeLike,
		model.RankingPeriodWeek,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": ranking,
	})
}

// GetLikeMonthRanking 获取点赞月榜
func (api *RankingAPI) GetLikeMonthRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	ranking, err := api.service.GetRanking(
		c.Request.Context(),
		model.RankingTypeLike,
		model.RankingPeriodMonth,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": ranking,
	})
}

// GetFavoriteWeekRanking 获取收藏周榜
func (api *RankingAPI) GetFavoriteWeekRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	ranking, err := api.service.GetRanking(
		c.Request.Context(),
		model.RankingTypeFavorite,
		model.RankingPeriodWeek,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": ranking,
	})
}

// GetFavoriteMonthRanking 获取收藏月榜
func (api *RankingAPI) GetFavoriteMonthRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	ranking, err := api.service.GetRanking(
		c.Request.Context(),
		model.RankingTypeFavorite,
		model.RankingPeriodMonth,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "查询成功",
		"data": ranking,
	})
}

// RebuildRankings 重建所有榜单（管理员接口）
func (api *RankingAPI) RebuildRankings(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("role")
	if !exists || (role != "admin" && role != "系统管理员") {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "权限不足"})
		return
	}

	err := api.service.RebuildAllRankings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "重建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "重建成功",
	})
}
