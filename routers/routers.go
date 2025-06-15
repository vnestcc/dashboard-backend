package routers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vnestcc/dashboard/docs"
	"github.com/vnestcc/dashboard/handlers"
	"github.com/vnestcc/dashboard/utils/values"
)

func LoadRoutes(r *gin.Engine) {
	apiRouter := r.Group("/api")
	cfg := values.GetConfig()
	if !cfg.Server.Prod {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		apiRouter.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	loadUserAuth(apiRouter)
	loadCompanies(apiRouter)
	loadManage(apiRouter)
	loadVCAuth(apiRouter)
	loadUser(apiRouter)

	apiRouter.GET("/ping", handlers.PingHandler)
}
