package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils/values"
)

var LoginCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 1,
	Revaluate:     true,
})

func generateJWT(id uint, role string) (string, error) {
	var cfg = values.GetConfig()
	claims := jwt.MapClaims{
		"id":   id,
		"role": role,
		"exp":  time.Now().Add(6 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Server.JWTSecret)
}

type userauthRequest struct {
	Email    string `json:"email" example:"example@vnest.org"`
	Password string `json:"password" example:"superstrongpassword"`
	Position string `json:"position" example:"founder"`
}

type authRequest struct {
	Email    string `json:"email" example:"example@vnest.org"`
	Password string `json:"password" example:"superstrongpassword"`
}

type successResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type failedResponse struct {
	Error string `json:"error"`
}

// UserSignupHandler godoc
// @Summary      User Signup
// @Description  Registers a new user with an email and password. The user is assigned a default role of "user".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    userauthRequest   true  "User Signup Input"
// @Success      200    {object}  successResponse
// @Failure      400    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/user/signup [post]
func UserSignupHandler(ctx *gin.Context) {
	var db = values.GetDB()
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Position string `json:"position"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user := models.User{
		Email:    input.Email,
		Password: input.Password,
		Position: input.Position,
		Role:     "user",
	}
	if err := db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}

// UserLoginHandler godoc
// @Summary      User Login
// @Description  Authenticates a user by email and password. Uses a cache lookup before querying the database. Returns a JWT token on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    authRequest  true  "User Login Input"
// @Success      200    {object}  successResponse
// @Failure      400    {object}  failedResponse
// @Failure      401    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/user/login [post]
func UserLoginHandler(ctx *gin.Context) {
	var db = values.GetDB()
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user models.User
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		user = value
		fmt.Println("fonud in cache")
	} else {
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		} else {
			fmt.Println("not in cache .. adding to cache")
			LoginCache.Set(user.Email, user)
		}
	}
	if ok, err := user.ComparePassword(input.Password); !ok || err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}

// VCSignupHandler godoc
// @Summary      VC Signup
// @Description  Registers a new VC user with an email and password. The user is assigned a default role of "vc".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    authRequest   true  "VC Signup Input"
// @Success      200    {object}  successResponse
// @Failure      400    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/vc/signup [post]
func VCSignupHandler(ctx *gin.Context) {
	var db = values.GetDB()
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user := models.User{
		Email:    input.Email,
		Password: input.Password,
		Role:     "vc",
	}
	if err := db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}

// VCLoginHandler godoc
// @Summary      VC Login
// @Description  Authenticates a user by email and password. Uses a cache lookup before querying the database. Returns a JWT token on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    authRequest  true  "VC Login Input"
// @Success      200    {object}  successResponse
// @Failure      400    {object}  failedResponse
// @Failure      401    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/vc/login [post]
func VCLoginHandler(ctx *gin.Context) {
	var db = values.GetDB()
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user models.User
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		user = value
		fmt.Println("fonud in cache")
	} else {
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		} else {
			fmt.Println("not in cache .. adding to cache")
			LoginCache.Set(user.Email, user)
		}
	}
	if ok, err := user.ComparePassword(input.Password); !ok || err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if !user.Approved {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "This account is still not approved"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}
