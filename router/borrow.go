package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitBorrowRouter(Router *gin.RouterGroup) {
	borrowRouter := Router.Group("borrow")
	borrowApi := v1.BorrowApi{}
	{
		borrowRouter.Use(middleware.JWTAuth())
		// 管理员接口
		borrowRouter.GET("getBorrowList", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), borrowApi.GetBorrowList)   // 获取借阅记录列表
		borrowRouter.POST("approve", middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian), borrowApi.ApproveBorrowRequest) // 审批借阅申请

		// 普通用户接口（所有登录用户都可以访问）
		borrowRouter.POST("borrowBook", borrowApi.BorrowBook)                   // 借书（普通用户提交申请，管理员直接借出）
		borrowRouter.POST("cancelBorrowRequest", borrowApi.CancelBorrowRequest) // 取消借阅申请
		borrowRouter.POST("returnBook", borrowApi.ReturnBook)                   // 还书
		borrowRouter.POST("renewBook", borrowApi.RenewBook)                     // 续借
		borrowRouter.GET("getMyBorrowList", borrowApi.GetMyBorrowList)          // 获取我的借阅记录
		borrowRouter.GET("getStatistics", borrowApi.GetBorrowStatistics)        // 获取我的借阅统计
	}
}
