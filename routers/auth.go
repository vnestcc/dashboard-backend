package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
)

func loadAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth")
	authRouter.POST("/forgot-password", handlers.ForgotPassword)
	authRouter.POST("/reset-password/:token", handlers.ResetPassword)

	loadUserAuth(authRouter)
	loadVCAuth(authRouter)
	loadAdminAuth(authRouter)
}

func loadUserAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/user")
	authRouter.POST("/signup", handlers.UserSignupHandler)
	authRouter.POST("/login", handlers.UserLoginHandler)
}

func loadVCAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/vc")
	authRouter.POST("/signup", handlers.VCSignupHandler)
	authRouter.POST("/login", handlers.VCLoginHandler)
}

func loadAdminAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/admin")
	authRouter.POST("/login", handlers.AdminLoginHandler)
}
