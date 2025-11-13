package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"

	"github.com/gin-gonic/gin"
)

func InitMessageRouter(Router *gin.RouterGroup) {
	messageRouter := Router.Group("message")
	messageApi := v1.MessageApi{}
	{
		messageRouter.Use(middleware.JWTAuth())
		messageRouter.GET("getMessages", messageApi.GetMessages)       // 获取消息列表
		messageRouter.GET("getUnreadCount", messageApi.GetUnreadCount) // 获取未读数量
		messageRouter.PUT("read/:id", messageApi.MarkAsRead)           // 标记已读
		messageRouter.PUT("readAll", messageApi.MarkAllAsRead)         // 全部标记已读
		messageRouter.DELETE("delete/:id", messageApi.DeleteMessage)   // 删除消息
	}
}
