package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitReservationRouter(Router *gin.RouterGroup) {
	reservationRouter := Router.Group("reservation")
	reservationApi := v1.ReservationApi{}
	{
		reservationRouter.Use(middleware.JWTAuth())
		// 普通用户接口
		reservationRouter.POST("create", reservationApi.CreateReservation)           // 创建预约
		reservationRouter.DELETE("cancel/:id", reservationApi.CancelReservation)     // 取消预约
		reservationRouter.GET("getMyReservations", reservationApi.GetMyReservations) // 获取我的预约

		// 管理员接口
		reservationRouter.GET("getReservationList", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), reservationApi.GetReservationList) // 获取预约列表
	}
}
