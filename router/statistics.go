package router

import (
	"bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"
	"github.com/gin-gonic/gin"
)

func InitStatisticsRouter(Router *gin.RouterGroup) {
	statisticsRouter := Router.Group("statistics")
	statisticsApi := v1.StatisticsApi{}
	{
		statisticsRouter.Use(middleware.JWTAuth())
		statisticsRouter.Use(middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian))
		statisticsRouter.GET("getStatistics", statisticsApi.GetStatistics)           // 获取统计信息
		statisticsRouter.GET("getBorrowStatistics", statisticsApi.GetBorrowStatistics) // 获取借阅统计
		statisticsRouter.GET("getPopularBooks", statisticsApi.GetPopularBooks)        // 获取热门图书
	}
}

