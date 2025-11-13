package router

import (
	v1 "bookadmin/api/v1"

	"github.com/gin-gonic/gin"
)

// InitRankingRouter 初始化榜单路由
func InitRankingRouter(router *gin.RouterGroup) {
	rankingAPI := v1.NewRankingAPI()
	
	rankingRouter := router.Group("/ranking")
	{
		// 通用榜单查询接口
		rankingRouter.GET("/list", rankingAPI.GetRanking)   // 统一榜单查询
		rankingRouter.GET("/query", rankingAPI.GetRanking)  // 兼容旧接口
		
		// 点赞榜
		rankingRouter.GET("/likes/week", rankingAPI.GetLikeWeekRanking)    // 点赞周榜
		rankingRouter.GET("/likes/month", rankingAPI.GetLikeMonthRanking)  // 点赞月榜
		
		// 收藏榜
		rankingRouter.GET("/favorites/week", rankingAPI.GetFavoriteWeekRanking)    // 收藏周榜
		rankingRouter.GET("/favorites/month", rankingAPI.GetFavoriteMonthRanking)  // 收藏月榜
		
		// 管理员接口
		rankingRouter.POST("/rebuild", rankingAPI.RebuildRankings) // 重建榜单
	}
}

