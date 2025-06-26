package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadManage(r *gin.RouterGroup) {
	manageRouter := r.Group("/manage")
	manageRouter.GET("/company/list", append(middleware.ModeratorMiddleware, handlers.ListCompany)...)
	manageRouter.GET("/company/:id", append(middleware.ModeratorMiddleware, handlers.GetCompanyByID)...)
	//	manageRouter.POST("/company/set", append(middleware.ModeratorMiddleware, handlers.SetCompanyParams)...)           // set visible/editable fields
	manageRouter.PUT("/company/edit/:id", append(middleware.ModeratorMiddleware, handlers.EditCompanyByID)...)
	manageRouter.DELETE("/company/delete/:id", append(middleware.ModeratorMiddleware, handlers.DeleteCompanyByID)...)

	manageRouter.GET("/vc/list", append(middleware.AdminMiddleware, handlers.GetVCList)...)
	manageRouter.PUT("/vc/:id/approve", append(middleware.AdminMiddleware, handlers.ApproveVC)...)
	manageRouter.PUT("/vc/:id/remove", append(middleware.AdminMiddleware, handlers.RemoveVC)...)
	manageRouter.DELETE("/vc/:id", append(middleware.AdminMiddleware, handlers.DeleteVC)...)

	manageRouter.GET("/users", append(middleware.AdminMiddleware, handlers.GetUserList)...)
	manageRouter.DELETE("/users/:id", append(middleware.AdminMiddleware, handlers.DeleteUserByID)...)
}
