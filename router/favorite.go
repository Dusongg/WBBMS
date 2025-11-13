package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"

	"github.com/gin-gonic/gin"
)

// InitFavoriteRouter 初始化收藏路由
func InitFavoriteRouter(router *gin.RouterGroup) {
	favoriteAPI := v1.NewFavoriteAPI()
	
	favoriteRouter := router.Group("/favorite")
	favoriteRouter.Use(middleware.JWTAuth()) // 需要登录
	{
		favoriteRouter.POST("/toggle/:bookId", favoriteAPI.ToggleFavorite)            // 切换收藏状态
		favoriteRouter.GET("/status/:bookId", favoriteAPI.GetFavoriteStatus)          // 查询收藏状态
		favoriteRouter.GET("/batch-status", favoriteAPI.BatchGetFavoriteStatus)       // 批量查询收藏状态
		favoriteRouter.GET("/list", favoriteAPI.GetUserFavoriteList)                  // 用户收藏列表
	}
}

