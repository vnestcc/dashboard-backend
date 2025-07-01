package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers/company"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadCompanies(r *gin.RouterGroup) {
	companyRouter := r.Group("/company")

	companyRouter.GET("/me", append(middleware.UserMiddleware, company.UserCompany)...)
	companyRouter.GET("/:id", middleware.JWTVerifyHandler, company.GetCompanyByID)
	companyRouter.GET("/list", company.ListCompany)
	companyRouter.GET("/quarters/:id", company.ListQuater)
	companyRouter.POST("/quarters/add", append(middleware.UserMiddleware, company.AddQuarter)...)
	companyRouter.POST("/create", append(middleware.UserMiddleware, company.CreateCompany)...)
	companyRouter.PUT("/edit", append(middleware.UserMiddleware, company.EditCompany)...)
	companyRouter.DELETE("/delete", append(middleware.UserMiddleware, company.DeleteCompany)...)
	companyRouter.POST("/join/:id", append(middleware.UserMiddleware, company.JoinCompany)...)
	companyRouter.GET("/perms/:id/visible")
	companyRouter.GET("/perms/editable")

	companyRouter.GET("/metrics/:id", middleware.JWTVerifyHandler, company.CompanyMetrics)
}
