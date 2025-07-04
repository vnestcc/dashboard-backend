package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	"github.com/vnestcc/dashboard/handlers/company"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadManage(r *gin.RouterGroup) {
	manageRouter := r.Group("/manage")
	manageRouter.GET("/company/list", append(middleware.ModeratorMiddleware, company.ListCompany)...)
	manageRouter.GET("/company/:id", append(middleware.ModeratorMiddleware, company.GetCompanyByID)...)
	//	manageRouter.POST("/company/set", append(middleware.ModeratorMiddleware, handlers.SetCompanyParams)...)           // set visible/editable fields
	manageRouter.PUT("/company/edit/:id", append(middleware.ModeratorMiddleware, company.EditCompanyByID)...)
	manageRouter.DELETE("/company/delete/:id", append(middleware.ModeratorMiddleware, company.DeleteCompanyByID)...)
	manageRouter.GET("/company/perms/:id/visible")
	manageRouter.GET("/company/perms/:id/editable")
	manageRouter.POST("/company/perms/:id/visible")
	manageRouter.POST("/company/perms/:id/editable")
	manageRouter.POST("/company/quarters/:id/new", append(middleware.ModeratorMiddleware, company.AllowQuarterByID)...)
	manageRouter.DELETE("/company/quarters/:id/remove", append(middleware.ModeratorMiddleware, company.RemoveQuarterByID)...)

	manageRouter.GET("/vc/list", append(middleware.AdminMiddleware, handlers.GetVCList)...)
	manageRouter.PUT("/vc/:id/approve", append(middleware.AdminMiddleware, handlers.ApproveVC)...)
	manageRouter.PUT("/vc/:id/remove", append(middleware.AdminMiddleware, handlers.RemoveVC)...)
	manageRouter.DELETE("/vc/:id", append(middleware.AdminMiddleware, handlers.DeleteVC)...)

	manageRouter.GET("/users", append(middleware.AdminMiddleware, handlers.GetUserList)...)
	manageRouter.DELETE("/users/:id", append(middleware.AdminMiddleware, handlers.DeleteUserByID)...)
}
