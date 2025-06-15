package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/users")
	userRouter.PUT("/", append(middleware.UserMiddleware, handlers.EditUser)...)
	userRouter.DELETE("/", append(middleware.UserMiddleware, handlers.DeleteUser)...)
	userRouter.GET("/me", append(middleware.UserMiddleware, handlers.UserMe)...)
}
