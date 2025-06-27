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
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	"github.com/vnestcc/dashboard/utils/values"
)

var LoginCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 1,
	CleanerMode:   cacher.CleaningCentral,
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
	Name     string `json:"name" example:"someone" binding:"required"`
	Email    string `json:"email" example:"example@vnest.org" binding:"required,email"`
	Password string `json:"password" example:"superstrongpassword" binding:"required,min=8"`
	Position string `json:"position" example:"founder" binding:"required"`
}

type vcauthRequest struct {
	Name     string `json:"name" example:"someone" binding:"required"`
	Email    string `json:"email" example:"example@vnest.org" binding:"required,email"`
	Password string `json:"password" example:"superstrongpassword" binding:"required,min=8"`
}

type authRequest struct {
	Email    string `json:"email" example:"example@vnest.org" binding:"required,email"`
	Password string `json:"password" example:"superstrongpassword" binding:"required,min=8"`
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
	db := values.GetDB()
	var input userauthRequest
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":   ctx.ClientIP(),
		"type": "audit",
	})
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "signup_attempt",
			"status": "failure",
			"reason": "invalid_json",
		}).Warn("Invalid signup input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user := models.User{
		Email:    input.Email,
		Password: input.Password,
		Position: input.Position,
		Name:     input.Name,
		Role:     "user",
	}
	if err := db.Create(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "signup_attempt",
			"status": "failure",
			"reason": "db_create_failed",
			"email":  user.Email,
			"error":  err.Error(),
		}).Error("User signup failed")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	token, err := generateJWT(user.ID, user.Role)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "signup_token_generation",
			"status":  "failure",
			"email":   user.Email,
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("JWT generation failed after signup")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "signup_attempt",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User signed up successfully")
	ctx.JSON(http.StatusOK, gin.H{"token": token})
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
	var input authRequest
	var user models.User
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "login_attempt",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid input during login")

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if cachedUser, ok := LoginCache.Get(input.Email); ok {
		ctx.Set("message", fmt.Sprintf("User %d loaded from cache", cachedUser.ID))
		user = cachedUser
		auditLog.WithFields(logrus.Fields{
			"event":   "login_cache_hit",
			"user_id": cachedUser.ID,
			"email":   cachedUser.Email,
			"ip":      ctx.ClientIP(),
		}).Info("User login cache hit")
	} else {
		if err := db.Where("email = ? AND role = 'user'", input.Email).First(&user).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":  "login_attempt",
				"email":  input.Email,
				"status": "failure",
				"reason": "user_not_found",
				"ip":     ctx.ClientIP(),
			}).Warn("Login failed - user not found")

			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		ctx.Set("message", fmt.Sprintf("User %d added to login cache", user.ID))
		LoginCache.Set(user.Email, user)
		auditLog.WithFields(logrus.Fields{
			"event":   "login_cache_miss",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("User loaded from DB and cached")
	}
	if err := user.ComparePassword(input.Password); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "login_attempt",
			"user_id": user.ID,
			"email":   user.Email,
			"status":  "failure",
			"reason":  "invalid_password",
			"ip":      ctx.ClientIP(),
		}).Warn("Login failed - invalid password")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token, err := generateJWT(user.ID, user.Role)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "token_generation",
			"user_id": user.ID,
			"email":   user.Email,
			"status":  "failure",
			"error":   err.Error(),
			"ip":      ctx.ClientIP(),
		}).Error("JWT generation failed")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "login_attempt",
		"user_id": user.ID,
		"email":   user.Email,
		"status":  "success",
		"ip":      ctx.ClientIP(),
	}).Info("User logged in successfully")
	LoginCache.Delete(user.Email)
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// VCSignupHandler godoc
// @Summary      VC Signup
// @Description  Registers a new VC user with an email and password. The user is assigned a default role of "vc".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    vcauthRequest   true  "VC Signup Input"
// @Success      200    {object}  successResponse
// @Failure      400    {object}  failedResponse
// @Failure      500    {object}  failedResponse
// @Router       /auth/vc/signup [post]
// NOTE: testing done
func VCSignupHandler(ctx *gin.Context) {
	var db = values.GetDB()
	var input vcauthRequest
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "vc_signup",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid VC signup input")

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user := models.User{
		Email:    input.Email,
		Password: input.Password,
		Role:     "vc",
		Name:     input.Name,
	}
	if err := db.Create(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "vc_signup",
			"status": "failure",
			"reason": "db_create_failed",
			"email":  user.Email,
			"ip":     ctx.ClientIP(),
			"error":  err.Error(),
		}).Error("VC signup DB creation failed")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the user"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "vc_signup",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
		"ip":      ctx.ClientIP(),
	}).Info("VC account submitted for approval")
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
	var input authRequest
	var user models.User
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "vc_login",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid VC login input")

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		user = value
		ctx.Set("message", fmt.Sprintf("User %d loaded from login cache", user.ID))

		auditLog.WithFields(logrus.Fields{
			"event":   "vc_login_cache_hit",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("VC login cache hit")
	} else {
		if err := db.Where("email = ? AND role = 'vc'", input.Email).First(&user).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":  "vc_login",
				"status": "failure",
				"reason": "user_not_found",
				"email":  input.Email,
				"ip":     ctx.ClientIP(),
			}).Warn("VC login failed - user not found")
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		ctx.Set("message", fmt.Sprintf("User %d added to cache", user.ID))
		LoginCache.Set(user.Email, user)
		auditLog.WithFields(logrus.Fields{
			"event":   "vc_login_cache_miss",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("VC user loaded from DB and cached")
	}
	if err := user.ComparePassword(input.Password); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "vc_login",
			"status":  "failure",
			"reason":  "invalid_password",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Warn("VC login failed - incorrect password")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if !user.Approved {
		auditLog.WithFields(logrus.Fields{
			"event":   "vc_login",
			"status":  "failure",
			"reason":  "not_approved",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Warn("VC login failed - account not approved")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "This account is still not approved"})
		return
	}
	token, err := generateJWT(user.ID, user.Role)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "vc_token_generation",
			"status":  "failure",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("VC JWT generation failed")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}
	LoginCache.Delete(user.Email)
	auditLog.WithFields(logrus.Fields{
		"event":   "vc_login",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
		"ip":      ctx.ClientIP(),
	}).Info("VC logged in successfully")
	ctx.JSON(http.StatusOK, gin.H{"token": token})
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
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "forgot_password",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid forgot password input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	otpSet := input.OTP != nil && *input.OTP != ""
	backupSet := input.BackupCode != nil && *input.BackupCode != ""
	if (otpSet && backupSet) || (!otpSet && !backupSet) {
		auditLog.WithFields(logrus.Fields{
			"event":  "forgot_password",
			"status": "failure",
			"reason": "invalid_auth_method_selection",
			"email":  input.Email,
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid OTP/Backup Code usage")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Provide either OTP or Backup Code, not both"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		ctx.Set("message", fmt.Sprintf("User %d loaded from login cache", value.ID))
		user = value
		auditLog.WithFields(logrus.Fields{
			"event":   "forgot_password_cache_hit",
			"status":  "info",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("User loaded from cache for password reset")
	} else {
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			ctx.Set("message", err.Error())
			auditLog.WithFields(logrus.Fields{
				"event":  "forgot_password",
				"status": "failure",
				"reason": "user_not_found",
				"email":  input.Email,
				"ip":     ctx.ClientIP(),
			}).Warn("User not found for forgot password")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
	}
	if otpSet && user.VerifyTOTP(*input.OTP) {
		ctx.Set("message", fmt.Sprintf("User %d reseting password using TOTP", user.ID))
		auditLog.WithFields(logrus.Fields{
			"event":   "forgot_password_auth",
			"status":  "success",
			"method":  "totp",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("Password reset authorized via TOTP")
	} else if backupSet && user.BackupCode == *input.BackupCode {
		ctx.Set("message", fmt.Sprintf("User %d reseting password using Backup code", user.ID))
		auditLog.WithFields(logrus.Fields{
			"event":   "forgot_password_auth",
			"status":  "success",
			"method":  "backup_code",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("Password reset authorized via Backup Code")
	} else {
		errMsg := "Invalid credentials"
		method := "unknown"
		if otpSet {
			errMsg = "Invalid OTP"
			method = "totp"
		} else if backupSet {
			errMsg = "Invalid Backup Code"
			method = "backup_code"
		}
		auditLog.WithFields(logrus.Fields{
			"event":   "forgot_password_auth",
			"status":  "failure",
			"reason":  "invalid_" + method,
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Warn("Failed password reset authentication")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		return
	}
	random := make([]byte, 20)
	_, err := rand.Read(random)
	if err != nil {
		ctx.Set("message", err.Error())
		auditLog.WithFields(logrus.Fields{
			"event":   "forgot_password",
			"status":  "failure",
			"reason":  "token_generation_failed",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to generate password reset token")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	token := hex.EncodeToString(random)
	ResetPasswordCache.Set(token, user)
	auditLog.WithFields(logrus.Fields{
		"event":   "forgot_password_token_issued",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
		"ip":      ctx.ClientIP(),
	}).Info("Password reset token successfully issued")
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
	auditLog := utils.Logger.WithField("type", "audit")
	if token == "" {
		auditLog.WithFields(logrus.Fields{
			"event":  "reset_password",
			"status": "failure",
			"reason": "missing_token",
			"ip":     ctx.ClientIP(),
		}).Warn("Reset password request missing token")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing reset token"})
		return
	}
	user, ok := ResetPasswordCache.Get(token)
	if !ok {
		auditLog.WithFields(logrus.Fields{
			"event":  "reset_password",
			"status": "failure",
			"reason": "invalid_or_expired_token",
			"token":  token,
			"ip":     ctx.ClientIP(),
		}).Warn("Reset password token invalid or expired")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	var input resetPasswordRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "reset_password",
			"status":  "failure",
			"reason":  "invalid_json",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid input during password reset")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := user.SetPassword(input.Password); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "reset_password",
			"status":  "failure",
			"reason":  "set_password_failed",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to hash new password")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}
	db := values.GetDB()
	if err := db.Model(&user).Update("password", user.Password).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "reset_password",
			"status":  "failure",
			"reason":  "db_update_failed",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to update password in DB")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	ResetPasswordCache.Delete(token)
	LoginCache.Delete(user.Email)
	auditLog.WithFields(logrus.Fields{
		"event":   "reset_password",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
		"ip":      ctx.ClientIP(),
	}).Info("Password reset successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}

// AdminLoginHandler godoc
// @Summary      Admin Login
// @Description  Authenticates an admin user by email and password. Uses a cache lookup before querying the database. Returns a JWT token on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    authRequest     true  "Admin Login Input"
// @Success      200      {object}  successResponse
// @Failure      400      {object}  failedResponse
// @Failure      401      {object}  failedResponse
// @Failure      500      {object}  failedResponse
// @Router       /auth/admin/login [post]
func AdminLoginHandler(ctx *gin.Context) {
	db := values.GetDB()
	var input authRequest
	var user models.User
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "admin_login",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid admin login input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if value, ok := LoginCache.Get(input.Email); ok {
		ctx.Set("message", fmt.Sprintf("Admin %d loaded from cache", value.ID))
		user = value
		auditLog.WithFields(logrus.Fields{
			"event":   "admin_login_cache_hit",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("Admin login cache hit")
	} else {
		if err := db.Where("email = ? AND role = ?", input.Email, "admin").First(&user).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":  "admin_login",
				"status": "failure",
				"reason": "user_not_found",
				"email":  input.Email,
				"ip":     ctx.ClientIP(),
			}).Warn("Admin login failed - user not found")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		ctx.Set("message", fmt.Sprintf("Admin %d added to login cache", user.ID))
		LoginCache.Set(user.Email, user)
		auditLog.WithFields(logrus.Fields{
			"event":   "admin_login_cache_miss",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Info("Admin loaded from DB and cached")
	}
	if err := user.ComparePassword(input.Password); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "admin_login",
			"status":  "failure",
			"reason":  "invalid_password",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
		}).Warn("Admin login failed - invalid password")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token, err := generateJWT(user.ID, user.Role)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "admin_token_generation",
			"status":  "failure",
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to generate admin JWT")
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}
	LoginCache.Delete(user.Email)
	auditLog.WithFields(logrus.Fields{
		"event":   "admin_login",
		"status":  "success",
		"user_id": user.ID,
		"email":   user.Email,
		"ip":      ctx.ClientIP(),
	}).Info("Admin logged in successfully")
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
