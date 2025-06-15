package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vnestcc/dashboard/models"
	"gorm.io/gorm"
)

var StartupCache = cacher.NewCacher[uint, models.Company](&cacher.NewCacherOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
})

type Claims struct {
	Id   uint
	Role string
	jwt.RegisteredClaims
}

// UserCompany godoc
// @Summary      Get current user's company
// @Description  Retrieves the company associated with the authenticated user
// @Tags         company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/me [get]
func UserCompany(ctx *gin.Context) {
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
	if err := db.Preload("StartUp").First(&user, claims.Id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartUp == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to any company"})
		return
	}
	startup := user.StartUp
	ctx.JSON(http.StatusOK, gin.H{
		"id":            startup.ID,
		"name":          startup.Name,
		"secret_code":   startup.SecretCode,
		"contact_name":  startup.ContactName,
		"contact_email": startup.ContactEmail,
	})
}

type quarterResponse struct {
	ID      uint   `json:"id" example:"1"`
	Quarter string `json:"quarter" example:"Q1"`
	Year    uint   `json:"year" example:"2025"`
	Date    string `json:"date,omitempty" example:"2025-04-01T00:00:00Z"`
}

// ListQuater godoc
// @Summary      List quarters by company
// @Description  Lists all quarters for the specified company
// @Tags         company
// @Produce      json
// @Param        id   path      int  true  "Company ID"
// @Success      200  {array}   quarterResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/quarters/{id} [get]
func ListQuater(ctx *gin.Context) {
	idStr := ctx.Param("id")
	companyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var company models.Company
	if cached, found := StartupCache.Get(uint(companyID)); found {
		company = cached
	} else {
		if err := db.First(&company, uint(companyID)).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch company"})
			return
		}
		StartupCache.Set(company.ID, company)
	}
	var quarters []models.Quarter
	if err := db.Where("company_id = ?", company.ID).Find(&quarters).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quarters"})
		return
	}
	ctx.JSON(http.StatusOK, quarters)
}

// ListCompany godoc
// @Summary      List all companies
// @Description  Retrieves a list of all companies available in the system
// @Tags         company
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/list [get]
func ListCompany(ctx *gin.Context) {
	var companies []models.Company
	result := make(map[uint]string)
	if err := db.Find(&companies).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the company list"})
		return
	}
	for i := range companies {
		StartupCache.Set(companies[i].ID, companies[i])
		result[companies[i].ID] = companies[i].Name
	}
	ctx.JSON(http.StatusOK, result)
}

// GetCompanyByID godoc
// @Summary      Get company details
// @Description  Returns the current user's company information
// @Tags         company
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/{id} [get]
// for admin .. show everything ..
// for user .. show based on isVisible
// for vc .. based on isVisible
// TODO: work
func GetCompanyByID(ctx *gin.Context) {
	// Extract and parse company ID
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)

	// Query parameters
	data := ctx.Query("data")
	quarter := ctx.Query("quarter")
	year := ctx.Query("year")

	// Whitelist for `data` query param
	allowedData := map[string]bool{
		"info": true, "finance": true, "market": true, "uniteconomics": true,
		"teamperf": true, "fund": true, "competitive": true, "operation": true,
		"risk": true, "additional": true, "self": true, "attachements": true,
	}

	if data != "" && !allowedData[data] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
		return
	}

	// Try getting from cache
	if cachedCompany, found := StartupCache.Get(companyID); found {
		ctx.JSON(http.StatusOK, gin.H{
			"company": cachedCompany,
			"data":    data,
			"quarter": quarter,
			"year":    year,
			"cached":  true,
		})
		return
	}

	// Not found in cache, fetch from DB
	var company models.Company
	if err := db.First(&company, companyID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	// Cache and return
	StartupCache.Set(companyID, company)
	ctx.JSON(http.StatusOK, gin.H{
		"company": company,
		"data":    data,
		"quarter": quarter,
		"year":    year,
		"cached":  false,
	})
}

type createCompanyRequest struct {
	Name         string `json:"name" binding:"required" example:"Acme Inc"`
	ContactName  string `json:"contact_name" binding:"required" example:"John Doe"`
	ContactEmail string `json:"contact_email" binding:"required,email" example:"john@acme.com"`
}

// CreateCompany godoc
// @Summary      Create a company
// @Description  Creates a new company for the user
// @Tags         company
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body      createCompanyRequest true  "Company details"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/create [post]
func CreateCompany(ctx *gin.Context) {
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(Claims)
	if !ok || claims.Role == "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot create companies"})
		return
	}
	var req struct {
		Name         string `json:"name" binding:"required"`
		ContactName  string `json:"contact_name" binding:"required"`
		ContactEmail string `json:"contact_email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	var user models.User
	if err := db.First(&user, claims.Id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartupID != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "User already belongs to a company"})
		return
	}
	newCompany := models.Company{
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
	}
	if err := db.Create(&newCompany).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Company with this contact email already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}
	if err := db.Model(&user).Update("startup_id", newCompany.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link company to user"})
		return
	}
	StartupCache.Set(newCompany.ID, newCompany)
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Company created successfully",
		"company_id": newCompany.ID,
	})
}

// EditCompany godoc
// @Summary      Edit company information
// @Description  Updates the existing company data
// @Security     BearerAuth
// @Tags         company
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/edit [put]
// for user .. based on isEditbale
// for admin .. everything
// TODO: work
func EditCompany(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "company edited"})
}

// TODO: work /manage/company/edit/id
func EditCompanyByID(ctx *gin.Context) {

}

// DeleteCompany godoc
// @Summary      Delete a company
// @Description  Deletes the user's company
// @Tags         company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /company/delete [delete]
func DeleteCompany(ctx *gin.Context) {
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
	if err := db.Preload("StartUp").First(&user, claims.Id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartUp == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to any company"})
		return
	}
	companyID := user.StartUp.ID
	if err := db.Delete(&models.Company{}, companyID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}
	if err := db.Model(&user).Update("startup_id", nil).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from user"})
		return
	}
	StartupCache.Delete(companyID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

// DeleteCompanyByID godoc
// @Summary      Admin delete company
// @Description  Deletes a company by its ID (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Company ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/delete/{id} [delete]
func DeleteCompanyByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)
	tx := db.Delete(&models.Company{}, companyID)
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}
	if tx.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Company does not exist"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("startup_id = ?", companyID).
		Update("startup_id", nil).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from users"})
		return
	}
	StartupCache.Delete(companyID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

type joinCompanyRequest struct {
	SecretCode string `json:"secret_code" binding:"required" example:"random hex"`
}

// JoinCompany godoc
// @Summary      Join a company
// @Description  Join company for the user
// @Tags         company
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int                true  "Company ID"
// @Param        body body      joinCompanyRequest true  "Secret Code"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/join/{id} [post]
func JoinCompany(ctx *gin.Context) {
	companyIDStr := ctx.Param("id")
	companyIDUint, err := strconv.ParseUint(companyIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(companyIDUint)
	var req struct {
		SecretCode string `json:"secret_code" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid secret code"})
		return
	}
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
	userID := claims.Id
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user.StartupID != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "User already belongs to a company"})
		return
	}
	var company models.Company
	if cachedCompany, found := StartupCache.Get(companyID); found {
		company = cachedCompany
	} else {
		if err := db.First(&company, companyID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company"})
			return
		}
		StartupCache.Set(company.ID, company)
	}
	if company.SecretCode != req.SecretCode {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret code"})
		return
	}
	if err := db.Model(&user).Update("startup_id", company.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join company"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully joined the company"})
}
