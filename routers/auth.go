package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/handlers"
)

func loadUserAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth/user")
	authRouter.POST("/signup", handlers.UserSignupHandler)
	authRouter.POST("/login", handlers.UserLoginHandler)
}

func loadVCAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth/vc")
	authRouter.POST("/signup", handlers.VCSignupHandler)
	authRouter.POST("/login", handlers.VCLoginHandler)
}
