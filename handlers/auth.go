package handlers

import (
	"crypto/rand"
	"encoding/hex"
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
	return token.SignedString([]byte(cfg.Server.JWTSecret))
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
// NOTE: testing done
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
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.Set("message", err.Error())
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
// NOTE: testing done
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
		ctx.Set("message", fmt.Sprintf("User %d loaded from cache", value.ID))
		user = value
	} else {
		if err := db.Where("email = ? AND role != 'vc'", input.Email).First(&user).Error; err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		} else {
			ctx.Set("message", fmt.Sprintf("User %d added to login cache", user.ID))
			LoginCache.Set(user.Email, user)
		}
	}
	if err := user.ComparePassword(input.Password); err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.Set("message", err.Error())
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
// NOTE: testing done
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
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Your account is submitted for approval"})
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
// NOTE: testing done
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
		ctx.Set("message", fmt.Sprintf("User %d loaded from login cache", value.ID))
		user = value
	} else {
		if err := db.Where("email = ? AND role != 'user'", input.Email).First(&user).Error; err != nil {
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		} else {
			ctx.Set("message", fmt.Sprintf("User %d added to cache", value.ID))
			LoginCache.Set(user.Email, user)
		}
	}
	if err := user.ComparePassword(input.Password); err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if !user.Approved {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "This account is still not approved"})
		return
	}
	if token, err := generateJWT(user.ID, user.Role); err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}

type forgotPasswordRequest struct {
	Email      string  `json:"email" example:"example@vnest.org" binding:"required,email"`
	OTP        *string `json:"otp" example:"112233"`
	BackupCode *string `json:"backup_code" example:"123456789012"`
}

type resetTokenResponse struct {
	ResetToken string `json:"reset_token" example:"abc123def456..."`
}

// ForgotPassword godoc
// @Summary      Forgot Password request
// @Description  Request to reset password using either OTP or Backup Code. Only one must be provided.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    forgotPasswordRequest  true  "Forgot Password Input"
// @Success      200    {object}  resetTokenResponse
// @Failure      400    {object}  failedResponse
// @Failure      401    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/forgot-password [post]
// NOTE: test done
func ForgotPassword(ctx *gin.Context) {
	var input forgotPasswordRequest
	var db = values.GetDB()
	var user models.User
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	otpSet := input.OTP != nil && *input.OTP != ""
	backupSet := input.BackupCode != nil && *input.BackupCode != ""
	if (otpSet && backupSet) || (!otpSet && !backupSet) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Provide either OTP or Backup Code, not both"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		ctx.Set("message", fmt.Sprintf("User %d loaded from login cache", value.ID))
		user = value
	} else {
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
	}
	if otpSet && user.VerifyTOTP(*input.OTP) {
		ctx.Set("message", fmt.Sprintf("User %d reseting password using TOTP", user.ID))
	} else if backupSet && user.BackupCode == *input.BackupCode {
		ctx.Set("message", fmt.Sprintf("User %d reseting password using Backup code", user.ID))
	} else {
		errMsg := "Invalid credentials"
		if otpSet {
			errMsg = "Invalid OTP"
		} else if backupSet {
			errMsg = "Invalid Backup Code"
		}
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		return
	}
	random := make([]byte, 20)
	_, err := rand.Read(random)
	if err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	token := hex.EncodeToString(random)
	ResetPasswordCache.Set(token, user)
	ctx.JSON(http.StatusOK, resetTokenResponse{
		ResetToken: token,
	})
}

type resetPasswordRequest struct {
	Password string `json:"password" example:"MyNewStrongPassword" binding:"required,min=8"`
}

// ResetPassword godoc
// @Summary      Reset Password
// @Description  Resets the user's password using the reset token issued after OTP/Backup Code verification. The token must be valid and not expired.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token    path     string                true  "Reset Token"
// @Param        request  body     resetPasswordRequest  true  "New Password"
// @Success      200      {object} map[string]string            "Password reset successful"
// @Failure      400      {object} failedResponse
// @Failure      401      {object} failedResponse
// @Failure      500      {object} failedResponse
// @Router       /auth/reset-password/{token} [post]
// NOTE: test done
func ResetPassword(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing reset token"})
		return
	}
	user, ok := ResetPasswordCache.Get(token)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	var input resetPasswordRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := user.SetPassword(input.Password); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}
	db := values.GetDB()
	if err := db.Model(&user).Update("password", user.Password).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	fmt.Println(user.Password) // TEST: remove
	ResetPasswordCache.Delete(token)
	LoginCache.Delete(user.Email)
	ctx.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}
