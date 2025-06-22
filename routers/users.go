package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

func loadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/users")
	userRouter.Use(middleware.UserMiddleware...)
	userRouter.PUT("/", handlers.EditUser)
	userRouter.DELETE("/", handlers.DeleteUser)
	userRouter.GET("/me", handlers.UserMe)
	userRouter.GET("/totp-qr", handlers.UserTOTP)
	userRouter.GET("/backup-code", handlers.UserBackupCode)
}
