package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitBookRouter(Router *gin.RouterGroup) {
	bookRouter := Router.Group("book")
	bookApi := v1.BookApi{}
	{
		// 公开接口：获取图书列表和详情（无需认证）
		bookRouter.GET("getBookList", bookApi.GetBookList) // 获取图书列表
		bookRouter.GET("getBook", bookApi.GetBook)         // 获取单个图书

		// 需要认证的接口
		bookRouter.Use(middleware.JWTAuth())
		bookRouter.Use(middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian))
		bookRouter.POST("createBook", bookApi.CreateBook)           // 新建图书
		bookRouter.DELETE("deleteBook", bookApi.DeleteBook)         // 删除图书
		bookRouter.DELETE("deleteBook/:id", bookApi.DeleteBookById) // 通过ID删除图书
		bookRouter.PUT("updateBook", bookApi.UpdateBook)            // 更新图书
	}
}
