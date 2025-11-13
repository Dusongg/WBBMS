package router

import (
	"bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"
	"github.com/gin-gonic/gin"
)

func InitSystemRouter(Router *gin.RouterGroup) {
	systemRouter := Router.Group("system")
	systemApi := v1.SystemApi{}
	{
		systemRouter.Use(middleware.JWTAuth())
		systemRouter.Use(middleware.RequireRole(model.RoleAdmin))
		systemRouter.GET("getUserList", systemApi.GetUserList)   // 获取用户列表
		systemRouter.POST("createUser", systemApi.CreateUser)    // 创建用户
		systemRouter.PUT("updateUser", systemApi.UpdateUser)     // 更新用户
		systemRouter.DELETE("deleteUser", systemApi.DeleteUser)  // 删除用户
	}
}

