package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadCompanies(r *gin.RouterGroup) {
	companyRouter := r.Group("/company")

	companyRouter.GET("/me", append(middleware.UserMiddleware, handlers.UserCompany)...)
	companyRouter.GET("/:id", middleware.JWTVerifyHandler, handlers.GetCompanyByID)
	companyRouter.GET("/list", handlers.ListCompany)
	companyRouter.GET("/quarters/:id", middleware.JWTVerifyHandler, handlers.ListQuater)
	companyRouter.POST("/create", append(middleware.UserMiddleware, handlers.CreateCompany)...)
	companyRouter.PUT("/edit", append(middleware.UserMiddleware, handlers.EditCompany)...)
	companyRouter.DELETE("/delete", append(middleware.UserMiddleware, handlers.DeleteCompany)...)
	companyRouter.POST("/join/:id", append(middleware.UserMiddleware, handlers.JoinCompany)...)
}
