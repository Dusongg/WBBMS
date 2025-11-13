package router

import (
	"bookadmin/api/v1"
	"bookadmin/middleware"
	"github.com/gin-gonic/gin"
)

func InitAuthRouter(Router *gin.RouterGroup) {
	authRouter := Router.Group("auth")
	authApi := v1.AuthApi{}
	{
	authRouter.POST("login", authApi.Login)           // 登录
	authRouter.POST("register", authApi.Register)     // 注册
		
	// 需要认证的接口
	authRouter.Use(middleware.JWTAuth())
	authRouter.GET("userInfo", authApi.GetUserInfo)   // 获取用户信息
	}
}

