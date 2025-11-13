package router

import (
	v1 "bookadmin/api/v1"
	"bookadmin/middleware"
	"bookadmin/model"

	"github.com/gin-gonic/gin"
)

func InitCategoryRouter(Router *gin.RouterGroup) {
	categoryRouter := Router.Group("category")
	categoryApi := v1.CategoryApi{}
	{
		// 公开接口：获取分类列表（无需认证）
		categoryRouter.GET("getCategoryList", categoryApi.GetCategoryList)

		// 需要认证的接口
		categoryRouter.Use(middleware.JWTAuth())
		categoryRouter.Use(middleware.RequireRole(model.RoleAdmin, model.RoleLibrarian))
		categoryRouter.POST("createCategory", categoryApi.CreateCategory)   // 新建分类
		categoryRouter.PUT("updateCategory", categoryApi.UpdateCategory)      // 更新分类
		categoryRouter.DELETE("deleteCategory", categoryApi.DeleteCategory)  // 删除分类
	}
}

