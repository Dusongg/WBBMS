package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"

	"github.com/gin-gonic/gin"
)

// InitLikeRouter 初始化点赞路由
func InitLikeRouter(router *gin.RouterGroup) {
	likeAPI := v1.NewLikeAPI()

	likeRouter := router.Group("/like")
	likeRouter.Use(middleware.JWTAuth()) // 需要登录
	{
		likeRouter.POST("/toggle/:bookId", likeAPI.ToggleLike)      // 切换点赞状态
		likeRouter.GET("/status/:bookId", likeAPI.GetLikeStatus)    // 查询点赞状态
		likeRouter.GET("/batch-status", likeAPI.BatchGetLikeStatus) // 批量查询点赞状态
		likeRouter.GET("/list", likeAPI.GetUserLikeList)            // 用户点赞列表
	}
}
