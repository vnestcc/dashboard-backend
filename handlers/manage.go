package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils/values"
)

type vcModel struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"someone"`
	Email    string `json:"email" example:"example@vnest.org"`
	Approved bool   `json:"approved" example:"false"`
}

// GetVCList godoc
// @Summary      Get list of VCs
// @Description  Retrieves a list of users with the VC role
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  []vcModel
// @Failure      500  {object}  failedResponse
// @Router       /manage/vc/list [get]
func GetVCList(ctx *gin.Context) {
	db := values.GetDB()
	var vc []models.User
	if err := db.Where("role = ?", "vc").Find(&vc).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the list of VCs"})
		return
	}
	result := make([]vcModel, 0, len(vc))
	for _, v := range vc {
		result = append(result, vcModel{
			ID:       v.ID,
			Name:     v.Name,
			Email:    v.Email,
			Approved: v.Approved,
		})
	}
	ctx.JSON(http.StatusOK, result)
}

// ApproveVC godoc
// @Summary      Approve a VC
// @Description  Sets the VC's approved field to true
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "VC ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  failedResponse
// @Failure      404  {object}  failedResponse
// @Failure      500  {object}  failedResponse
// @Router       /manage/vc/{id}/approve [put]
func ApproveVC(ctx *gin.Context) {
	var db = values.GetDB()
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	result := db.Model(&models.User{}).Where("id = ?", uint(id)).Update("approved", true)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve VC"})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "VC approved"})
}

// RemoveVC godoc
// @Summary      Unapprove a VC
// @Description  Sets the VC's approved field to false
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "VC ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  failedResponse
// @Failure      404  {object}  failedResponse
// @Failure      500  {object}  failedResponse
// @Router       /manage/vc/{id}/remove [put]
func RemoveVC(ctx *gin.Context) {
	var db = values.GetDB()
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result := db.Model(&models.User{}).Where("id = ?", uint(id)).Update("approved", false)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove VC approval"})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "VC approval removed"})
}

// DeleteVC godoc
// @Summary      Delete a VC
// @Description  Deletes the VC from the database by ID
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "VC ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  failedResponse
// @Failure      404  {object}  failedResponse
// @Failure      500  {object}  failedResponse
// @Router       /manage/vc/{id} [delete]
func DeleteVC(ctx *gin.Context) {
	db := values.GetDB()
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	result := db.Unscoped().Where("id = ? AND role = ?", uint(id), "vc").Delete(&models.User{})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete VC"})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "VC does not exist"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "VC deleted"})
}

// DeleteUserByID godoc
// @Summary      Delete a User
// @Description  Deletes a User by ID (hard delete)
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  failedResponse
// @Failure      404  {object}  failedResponse
// @Failure      500  {object}  failedResponse
// @Router       /manage/users/{id} [delete]
func DeleteUserByID(ctx *gin.Context) {
	db := values.GetDB()
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	result := db.Unscoped().Where("id = ? AND role = ?", uint(id), "user").Delete(&models.User{})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

type userModel struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"john@example.com"`
	IsDeleted bool   `json:"is_deleted" example:"false"`
	StartUpID *uint  `json:"startup_id,omitempty" example:"42"`
}

// GetUserList godoc
// @Summary      Get all users
// @Description  Returns a list of all users with the "user" role
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {array}   userModel
// @Failure      500  {object}  failedResponse
// @Router       /manage/users [get]
func GetUserList(ctx *gin.Context) {
	db := values.GetDB()
	var users []models.User
	if err := db.Where("role = ?", "user").Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the list of Users"})
		return
	}
	result := make([]userModel, 0, len(users))
	for _, v := range users {
		result = append(result, userModel{
			ID:        v.ID,
			Name:      v.Name,
			Email:     v.Email,
			StartUpID: v.StartupID,
			IsDeleted: v.DeletedAt.Valid,
		})
	}
	ctx.JSON(http.StatusOK, result)
}
