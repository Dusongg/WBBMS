package router

import (
	"bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"
	"github.com/gin-gonic/gin"
)

func InitReaderRouter(Router *gin.RouterGroup) {
	readerRouter := Router.Group("reader")
	readerApi := v1.ReaderApi{}
	{
		readerRouter.Use(middleware.JWTAuth())
		readerRouter.GET("getReaderList", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), readerApi.GetReaderList) // 获取读者列表
		readerRouter.GET("getReader", readerApi.GetReader)                                                                        // 获取读者信息
		readerRouter.PUT("updateReaderStatus", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), readerApi.UpdateReaderStatus) // 更新读者状态（审核）
		readerRouter.PUT("updateReader", readerApi.UpdateReader)                                                                  // 更新读者信息
	}
}

