package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
)

func loadUserAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth/user")
	authRouter.POST("/signup", handlers.UserSignupHandler)
	authRouter.POST("/login", handlers.UserLoginHandler)
	authRouter.POST("/forgot-password", func(ctx *gin.Context) {}) // send totp and then use it to verify and send a token as response
	authRouter.POST("/reset-password", func(ctx *gin.Context) {})  // using the given token update the password
}

func loadVCAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth/vc")
	authRouter.POST("/signup", handlers.VCSignupHandler)
	authRouter.POST("/login", handlers.VCLoginHandler)
	authRouter.POST("/forgot-password", func(ctx *gin.Context) {}) // send totp and then use it to verify and send a token as response
	authRouter.POST("/reset-password", func(ctx *gin.Context) {})  // using the given token update the password
}
