package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils/values"
)

type editUserRequest struct {
	Name     string `json:"name" example:"John Doe" binding:"required"`
	Position string `json:"position" example:"CTO" binding:"required"`
}

var UserCache = cacher.NewCacher[uint, models.User](&cacher.NewCacherOpts{
	Revaluate:     true,
	CleanInterval: 1 * time.Hour,
	TimeToLive:    3 * time.Minute,
	CleanerMode:   cacher.CleaningCentral,
})

// EditUser godoc
// @Summary      Edit user profile
// @Description  Allows a user to edit only their name and position
// @Tags         user
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body     editUserRequest true "Edit user name and position"
// @Success      200  {object} map[string]string
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /users [put]
// NOTE: testing done
func EditUser(ctx *gin.Context) {
	var db = values.GetDB()
	claimsAny, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsAny.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}
	var req editUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("id = ?", claims.ID).
		Updates(map[string]any{
			"name":     req.Name,
			"position": req.Position,
		}).Error; err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	UserCache.Delete(claims.ID)
	ctx.Set("message", fmt.Sprintf("Deleted User %d from cache", claims.ID))
	ctx.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

// DeleteUser godoc
// @Summary      Delete user account
// @Description  Deletes the current authenticated user
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /users [delete]
// NOTE: testing done
func DeleteUser(ctx *gin.Context) {
	var db = values.GetDB()
	claimsAny, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsAny.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}
	if err := db.Delete(&models.User{}, claims.ID).Error; err != nil {
		ctx.Set("message", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	UserCache.Delete(claims.ID)
	ctx.Set("message", fmt.Sprintf("Deleted User %d from cache", claims.ID))
	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

type userMeResponse struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"Alice"`
	Position string `json:"position" example:"CEO"`
	Email    string `json:"email" example:"alice@example.com"`
	Approved bool   `json:"approved" example:"false"`
}

// UserMe godoc
// @Summary      Get current user info
// @Description  Returns details of the authenticated user
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  userMeResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/me [get]
// NOTE: testing done
func UserMe(ctx *gin.Context) {
	var db = values.GetDB()
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if val, ok := UserCache.Get(claims.ID); ok {
		ctx.Set("message", fmt.Sprintf("Loaded User %d from cache", claims.ID))
		user = val
	} else {
		if err := db.First(&user, claims.ID).Error; err != nil {
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		} else {
			ctx.Set("message", fmt.Sprintf("Added User %d to cache", claims.ID))
			UserCache.Set(claims.ID, user)
		}
	}
	ctx.JSON(http.StatusOK, userMeResponse{
		ID:       user.ID,
		Name:     user.Name,
		Position: user.Position,
		Email:    user.Email,
		Approved: user.Approved,
	})
}

// UserTOTP godoc
// @Summary      Get TOTP QR Code
// @Description  Returns a QR code image for enabling TOTP for the authenticated user
// @Tags         user
// @Security     BearerAuth
// @Produce      png
// @Success      200  {file}  png
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/totp-qr [get]
// NOTE: testing done
func UserTOTP(ctx *gin.Context) {
	var db = values.GetDB()
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if val, ok := UserCache.Get(claims.ID); ok {
		ctx.Set("message", fmt.Sprintf("Loaded User %d from cache", claims.ID))
		user = val
	} else {
		if err := db.First(&user, claims.ID).Error; err != nil {
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		} else {
			ctx.Set("message", fmt.Sprintf("Added User %d to cache", claims.ID))
			UserCache.Set(claims.ID, user)
		}
	}
	totp_url, _ := user.TOTPUrl()
	png, err := qrcode.Encode(totp_url, qrcode.Medium, 256)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}
	ctx.Header("Content-Type", "image/png")
	ctx.Writer.Write(png)
}

// UserBackupCode godoc
// @Summary      Get TOTP Backup Code
// @Description  Returns the TOTP backup code for the authenticated user
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/backup-code [get]
// NOTE: testing done
func UserBackupCode(ctx *gin.Context) {
	var db = values.GetDB()
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if val, ok := UserCache.Get(claims.ID); ok {
		ctx.Set("message", fmt.Sprintf("Loaded User %d from cache", claims.ID))
		user = val
	} else {
		if err := db.First(&user, claims.ID).Error; err != nil {
			ctx.Set("message", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		} else {
			ctx.Set("message", fmt.Sprintf("Added User %d to cache", claims.ID))
			UserCache.Set(claims.ID, user)
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"backup_code": user.BackupCode})
}
