package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitBlacklistRouter(Router *gin.RouterGroup) {
	blacklistRouter := Router.Group("blacklist")
	blacklistApi := v1.BlacklistApi{}
	{
		blacklistRouter.Use(middleware.JWTAuth())
		// 普通用户接口
		blacklistRouter.GET("getMyStatus", blacklistApi.GetMyBlacklistStatus)  // 获取我的黑名单状态
		
		// 管理员接口
		blacklistRouter.GET("getBlacklistList", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), blacklistApi.GetBlacklistList)    // 获取黑名单列表
		blacklistRouter.POST("add", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), blacklistApi.AddToBlacklist)                  // 添加黑名单
		blacklistRouter.POST("remove/:id", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), blacklistApi.RemoveFromBlacklist)      // 解除黑名单
	}
}

