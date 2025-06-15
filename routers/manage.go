package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadManage(r *gin.RouterGroup) {
	manageRouter := r.Group("/manage")
	//manageRouter.GET("/company/list", append(middleware.ModeratorMiddleware, handlers.GetCompaniesList)...)   // list all companies
	//manageRouter.GET("/company/:id", append(middleware.ModeratorMiddleware, handlers.GetCompanyByIDAdmin)...) // get details of a company by id (for admin)
	//manageRouter.POST("/company/set", append(middleware.ModeratorMiddleware, handlers.SetCompanyParams)...)   // set fields visible or editable
	manageRouter.DELETE("/company/delete/:id", append(middleware.ModeratorMiddleware, handlers.DeleteCompanyByID)...)
	manageRouter.PUT("/company/edit/:id", append(middleware.ModeratorMiddleware, handlers.EditCompanyByID)...)

	manageRouter.GET("/vc/list", append(middleware.AdminMiddleware, handlers.GetVCList)...)
	manageRouter.PUT("/vc/:id/approve", append(middleware.AdminMiddleware, handlers.ApproveVC)...)
	manageRouter.PUT("/vc/:id/remove", append(middleware.AdminMiddleware, handlers.RemoveVC)...)
	manageRouter.DELETE("/vc/:id", append(middleware.AdminMiddleware, handlers.DeleteVC)...)
}
