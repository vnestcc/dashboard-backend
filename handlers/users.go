package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/models"
)

type editUserRequest struct {
	Name     string `json:"name" example:"John Doe" binding:"required"`
	Position string `json:"position" example:"CTO" binding:"required"`
}

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
// @Router       /user/edit [put]
func EditUser(ctx *gin.Context) {
	claimsAny, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsAny.(Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}
	var req editUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("id = ?", claims.ID).
		Updates(map[string]any{
			"name":     req.Name,
			"position": req.Position,
		}).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
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
// @Router       /user/delete [delete]
func DeleteUser(ctx *gin.Context) {
	claimsAny, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsAny.(Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}
	if err := db.Delete(&models.User{}, claims.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

type userMeResponse struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"Alice"`
	Position string `json:"position" example:"CEO"`
	Email    string `json:"email" example:"alice@example.com"`
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
// @Router       /user/me [get]
func UserMe(ctx *gin.Context) {
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if err := db.First(&user, claims.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	ctx.JSON(http.StatusOK, userMeResponse{
		ID:       user.ID,
		Name:     user.Name,
		Position: user.Position,
		Email:    user.Email,
	})
}
