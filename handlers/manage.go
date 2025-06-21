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
	var db = values.GetDB()
	var vc []models.User
	role := "vc"
	if err := db.Where("role = ?", role).Find(&vc).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the list of vc"})
		return
	}
	ctx.JSON(http.StatusOK, vc)
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
// @Description  Deletes the VC from the database
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
	var db = values.GetDB()
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result := db.Where("id = ?", uint(id)).Delete(&models.User{})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete VC"})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "VC deleted"})
}
