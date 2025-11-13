package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitFineRouter(Router *gin.RouterGroup) {
	fineRouter := Router.Group("fine")
	fineApi := v1.FineApi{}
	{
		fineRouter.Use(middleware.JWTAuth())
		// 普通用户接口
		fineRouter.GET("getMyFines", fineApi.GetMyFines)  // 获取我的罚款
		fineRouter.POST("pay", fineApi.PayFine)           // 支付罚款
		
		// 管理员接口
		fineRouter.GET("getFineList", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), fineApi.GetFineList)  // 获取罚款列表
		fineRouter.POST("waive/:id", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), fineApi.WaiveFine)     // 豁免罚款
	}
}

