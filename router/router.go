package router

import (
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	Router := gin.Default()

	// 跨域配置
	Router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API路由组
	apiRouter := Router.Group("api")
	{
		// 认证相关（无需JWT）
		InitAuthRouter(apiRouter)

		// 图书管理（需要JWT，部分接口需要管理员权限）
		InitBookRouter(apiRouter)

		// 分类管理
		InitCategoryRouter(apiRouter)

		// 读者管理
		InitReaderRouter(apiRouter)

		// 借还管理
		InitBorrowRouter(apiRouter)

		// 统计查询
		InitStatisticsRouter(apiRouter)

		// 系统管理
		InitSystemRouter(apiRouter)

		// 点赞功能
		InitLikeRouter(apiRouter)

		// 收藏功能
		InitFavoriteRouter(apiRouter)

		// 榜单功能
		InitRankingRouter(apiRouter)

		// 预约管理
		InitReservationRouter(apiRouter)

		// 罚款管理
		InitFineRouter(apiRouter)

		// 黑名单管理
		InitBlacklistRouter(apiRouter)

		// 消息管理
		InitMessageRouter(apiRouter)
	}

	return Router
}
