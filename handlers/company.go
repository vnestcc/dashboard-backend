package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/models"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
	"github.com/vnestcc/dashboard/utils/values"
	"gorm.io/gorm"
)

var StartupCache = cacher.NewCacher[uint, models.Company](&cacher.NewCacherOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
})

type Claims middleware.Claims

// UserCompany godoc
// @Summary      Get current user's company
// @Description  Retrieves the company associated with the authenticated user
// @Tags         company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/me [get]
func UserCompany(ctx *gin.Context) {
	var db = values.GetDB()
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
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
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
	var db = values.GetDB()
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
	var db = values.GetDB()
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

var QuarterCache = cacher.NewCacher[string, models.Quarter](&cacher.NewCacherOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
})

// GetCompanyByID godoc
// @Summary      Get company details
// @Description  Returns the current user's company information, including selectable related data sets
// @Tags         company
// @Produce      json
// @Param        id       path   int     true  "Company ID"
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/{id} [get]
// OPTIMIZE: there's too much object creations. Need fixing by hand written sql queries
func GetCompanyByID(ctx *gin.Context) {
	var db = values.GetDB()
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)
	var company models.Company
	if err := db.Where("id = ?", companyID).Find(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find company"})
		return
	}
	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Only one 'data' query parameter allowed"})
		return
	}
	data := ""
	if len(dataVals) == 1 {
		data = dataVals[0]
	}
	quarter := ctx.Query("quarter")
	yearStr := ctx.Query("year")
	allowedData := map[string]string{
		"info":          "",
		"finance":       "FinancialHealths",
		"market":        "MarketTractions",
		"uniteconomics": "UnitEconomics",
		"teamperf":      "TeamPerformances",
		"fund":          "FundraisingStatuses",
		"competitive":   "CompetitiveLandscapes",
		"operation":     "OperationalEfficiencies",
		"risk":          "RiskManagements",
		"additional":    "AdditionalInfos",
		"self":          "SelfAssessments",
		"attachements":  "Attachments",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
	cacheKey := fmt.Sprintf("%d_%s_%d", companyID, quarter, year)
	var quarterObj models.Quarter
	cached, found := QuarterCache.Get(cacheKey)
	if found {
		quarterObj = cached
	} else {
		err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).
			First(&quarterObj).Error
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
			return
		}
		QuarterCache.Set(cacheKey, quarterObj)
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
	var user models.User
	if err := db.First(&user, claims.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	full_access := (claims.Role == "admin") || (company.ID == *user.StartupID)
	switch data {
	case "", "info":
		ctx.JSON(http.StatusOK, gin.H{
			"company_id":            company.ID,
			"company_name":          company.Name,
			"company_contact_name":  company.ContactName,
			"company_contact_email": company.ContactEmail,
		})
	case "finance":
		var financialHealths []models.FinancialHealth
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&financialHealths).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load FinancialHealths"})
			return
		}
		var filtered []map[string]any
		for i := range financialHealths {
			filtered = append(filtered, financialHealths[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":        quarterObj.ID,
			"financial_healths": filtered,
		})
	case "market":
		var marketTractions []models.MarketTraction
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&marketTractions).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load MarketTractions"})
			return
		}
		var filtered []map[string]any
		for i := range marketTractions {
			filtered = append(filtered, marketTractions[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":       quarterObj.ID,
			"market_tractions": filtered,
		})
	case "uniteconomics":
		var unitEconomics []models.UnitEconomics
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&unitEconomics).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load UnitEconomics"})
			return
		}
		var filtered []map[string]any
		for i := range unitEconomics {
			filtered = append(filtered, unitEconomics[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":     quarterObj.ID,
			"unit_economics": filtered,
		})
	case "teamperf":
		var teamPerformances []models.TeamPerformance
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&teamPerformances).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load TeamPerformances"})
			return
		}
		var filtered []map[string]any
		for i := range teamPerformances {
			filtered = append(filtered, teamPerformances[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":        quarterObj.ID,
			"team_performances": filtered,
		})
	case "fund":
		var fundraisingStatuses []models.FundraisingStatus
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&fundraisingStatuses).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load FundraisingStatuses"})
			return
		}
		var filtered []map[string]any
		for i := range fundraisingStatuses {
			filtered = append(filtered, fundraisingStatuses[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":           quarterObj.ID,
			"fundraising_statuses": filtered,
		})
	case "competitive":
		var competitiveLandscapes []models.CompetitiveLandscape
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&competitiveLandscapes).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load CompetitiveLandscapes"})
			return
		}
		var filtered []map[string]any
		for i := range competitiveLandscapes {
			filtered = append(filtered, competitiveLandscapes[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":             quarterObj.ID,
			"competitive_landscapes": filtered,
		})
	case "operation":
		var operationalEfficiencies []models.OperationalEfficiency
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&operationalEfficiencies).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load OperationalEfficiencies"})
			return
		}
		var filtered []map[string]any
		for i := range operationalEfficiencies {
			filtered = append(filtered, operationalEfficiencies[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":               quarterObj.ID,
			"operational_efficiencies": filtered,
		})
	case "risk":
		var riskManagements []models.RiskManagement
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&riskManagements).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load RiskManagements"})
			return
		}
		var filtered []map[string]any
		for i := range riskManagements {
			filtered = append(filtered, riskManagements[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":      quarterObj.ID,
			"risk_management": filtered,
		})
	case "additional":
		var additionalInfos []models.AdditionalInfo
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&additionalInfos).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load AdditionalInfos"})
			return
		}
		var filtered []map[string]any
		for i := range additionalInfos {
			filtered = append(filtered, additionalInfos[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":      quarterObj.ID,
			"additional_info": filtered,
		})
	case "self":
		var selfAssessments []models.SelfAssessment
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&selfAssessments).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load SelfAssessments"})
			return
		}
		var filtered []map[string]any
		for i := range selfAssessments {
			filtered = append(filtered, selfAssessments[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":      quarterObj.ID,
			"self_assessment": filtered,
		})
	case "attachements":
		var attachments []models.Attachment
		err := db.Where("quarter_id = ?", quarterObj.ID).Find(&attachments).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load Attachments"})
			return
		}
		var filtered []map[string]any
		for i := range attachments {
			filtered = append(filtered, attachments[i].VisibilityFilter(full_access))
		}
		ctx.JSON(http.StatusOK, gin.H{
			"quarter_id":  quarterObj.ID,
			"attachments": filtered,
		})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
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
	var db = values.GetDB()
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
	if err := db.First(&user, claims.ID).Error; err != nil {
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

// GetCompanyByIDAdmin godoc
// @Summary      Get company details (Admin)
// @Description  Returns the specified company's information, including selectable related data sets (admin only).
// @Tags         company
// @Produce      json
// @Param        id       path   int     true  "Company ID"
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/{id} [get]
func GetCompanyByIDAdmin(ctx *gin.Context) {

}

// EditCompany godoc
// @Summary      Edit company information
// @Description  Updates the existing company data. If `data=info` or omitted, updates company name, contact name, and contact email. Otherwise, allows versioned updates for specific company data types (such as finance, market, uniteconomics, etc) for a given quarter and year. The allowed types are: finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements. All data modifications are subject to field-level editability checks based on the current IsEditable mask for each record.
// @Security     BearerAuth
// @Tags         company
// @Accept       json
// @Produce      json
// @Param        data   query     string  false  "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter query    string  false  "Quarter name (e.g. Q1, Q2, Q3, Q4). Required unless data=info" Enum(Q1,Q2,Q3,Q4)
// @Param        year    query    int  false  "Year (e.g. 2024). Required unless data=info"
// @Param        body    body     object  true   "Payload matching the type of data being edited"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string  "Invalid request or body"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Not permitted by edit mask"
// @Failure      404  {object}  map[string]string  "Company or quarter not found"
// @Failure      500  {object}  map[string]string  "Server/database error"
// @Router       /company/edit [put]
func EditCompany(ctx *gin.Context) {
	var db = values.GetDB()
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
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	companyID := user.StartupID

	var company models.Company
	if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
		return
	}

	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Only one 'data' query parameter allowed"})
		return
	}
	data := ""
	if len(dataVals) == 1 {
		data = dataVals[0]
	}
	quarter := ctx.Query("quarter")
	yearStr := ctx.Query("year")
	allowedData := map[string]string{
		"info":          "",
		"finance":       "FinancialHealths",
		"market":        "MarketTractions",
		"uniteconomics": "UnitEconomics",
		"teamperf":      "TeamPerformances",
		"fund":          "FundraisingStatuses",
		"competitive":   "CompetitiveLandscapes",
		"operation":     "OperationalEfficiencies",
		"risk":          "RiskManagements",
		"additional":    "AdditionalInfos",
		"self":          "SelfAssessments",
		"attachements":  "Attachments",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
	if data == "" || data == "info" {
		var req struct {
			Name         *string `json:"name"`
			ContactName  *string `json:"contact_name"`
			ContactEmail *string `json:"contact_email"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Name != nil {
			company.Name = *req.Name
		}
		if req.ContactName != nil {
			company.ContactName = *req.ContactName
		}
		if req.ContactEmail != nil {
			company.ContactEmail = *req.ContactEmail
		}
		if err := db.Save(&company).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Company updated successfully", "company": company})
		return
	}
	var quarterObj models.Quarter
	if err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).
		First(&quarterObj).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
		return
	}
	preloadField := allowedData[data]
	if preloadField == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No editable data specified"})
		return
	}
	switch data {
	case "finance":
		var req []models.FinancialHealth
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.FinancialHealth{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FinancialHealths versioned and inserted"})
	case "market":
		var req []models.MarketTraction
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.MarketTraction{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "MarketTractions versioned and inserted"})
	case "uniteconomics":
		var req []models.UnitEconomics
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.UnitEconomics{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "UnitEconomics versioned and inserted"})
	case "teamperf":
		var req []models.TeamPerformance
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.TeamPerformance{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "TeamPerformances versioned and inserted"})
	case "fund":
		var req []models.FundraisingStatus
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.FundraisingStatus{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FundraisingStatuses versioned and inserted"})
	case "competitive":
		var req []models.CompetitiveLandscape
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.CompetitiveLandscape{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "CompetitiveLandscapes versioned and inserted"})
	case "operation":
		var req []models.OperationalEfficiency
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.OperationalEfficiency{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "OperationalEfficiencies versioned and inserted"})
	case "risk":
		var req []models.RiskManagement
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.RiskManagement{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "RiskManagements versioned and inserted"})
	case "additional":
		var req []models.AdditionalInfo
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.AdditionalInfo{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "AdditionalInfos versioned and inserted"})
	case "self":
		var req []models.SelfAssessment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.SelfAssessment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "SelfAssessments versioned and inserted"})
	case "attachements":
		var req []models.Attachment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			if err := newObj.EditableFilter(); err != nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			var maxVersion int
			db.Model(&models.Attachment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Attachments versioned and inserted"})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
}

// EditCompanyByID godoc
// @Summary      Edit company details (Admin, versioned insert)
// @Description  Allows admin to insert new versioned data for company or related quarter data. If `data=info` or omitted, updates company name, contact name, and contact email. Otherwise, allows versioned updates for specific company data types (such as finance, market, uniteconomics, etc) for a given quarter and year. The allowed types are: finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements. All data modifications will insert a new version for the specified quarter and year.
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "Company ID"
// @Param        data    query     string  false "Type of company data to edit (info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)"
// @Param        quarter query     string  false "Quarter name (e.g. Q1, Q2, Q3, Q4). Required unless data=info" Enum(Q1,Q2,Q3,Q4)
// @Param        year    query     int     false "Year (e.g. 2024). Required unless data=info"
// @Param        body    body      object  true  "Payload matching the type of data being edited"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string  "Invalid request or body"
// @Failure      404  {object}  map[string]string  "Company or quarter not found"
// @Failure      500  {object}  map[string]string  "Server/database error"
// @Router       /manage/company/edit/{id} [put]
// OPTIMIZE: whatever is written here is not meant for production. I pray for the server :pray:
func EditCompanyByID(ctx *gin.Context) {
	var db = values.GetDB()
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)

	var company models.Company
	if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
		return
	}

	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Only one 'data' query parameter allowed"})
		return
	}
	data := ""
	if len(dataVals) == 1 {
		data = dataVals[0]
	}
	quarter := ctx.Query("quarter")
	yearStr := ctx.Query("year")
	allowedData := map[string]string{
		"info":          "",
		"finance":       "FinancialHealths",
		"market":        "MarketTractions",
		"uniteconomics": "UnitEconomics",
		"teamperf":      "TeamPerformances",
		"fund":          "FundraisingStatuses",
		"competitive":   "CompetitiveLandscapes",
		"operation":     "OperationalEfficiencies",
		"risk":          "RiskManagements",
		"additional":    "AdditionalInfos",
		"self":          "SelfAssessments",
		"attachements":  "Attachments",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
	if data == "" || data == "info" {
		var req struct {
			Name         *string `json:"name"`
			ContactName  *string `json:"contact_name"`
			ContactEmail *string `json:"contact_email"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Name != nil {
			company.Name = *req.Name
		}
		if req.ContactName != nil {
			company.ContactName = *req.ContactName
		}
		if req.ContactEmail != nil {
			company.ContactEmail = *req.ContactEmail
		}
		if err := db.Save(&company).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Company updated successfully", "company": company})
		return
	}
	var quarterObj models.Quarter
	if err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).
		First(&quarterObj).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
		return
	}
	preloadField := allowedData[data]
	if preloadField == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No editable data specified"})
		return
	}
	switch data {
	case "finance":
		var req []models.FinancialHealth
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.FinancialHealth{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FinancialHealths versioned and inserted"})
	case "market":
		var req []models.MarketTraction
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.MarketTraction{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "MarketTractions versioned and inserted"})
	case "uniteconomics":
		var req []models.UnitEconomics
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.UnitEconomics{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "UnitEconomics versioned and inserted"})
	case "teamperf":
		var req []models.TeamPerformance
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.TeamPerformance{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "TeamPerformances versioned and inserted"})
	case "fund":
		var req []models.FundraisingStatus
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.FundraisingStatus{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FundraisingStatuses versioned and inserted"})
	case "competitive":
		var req []models.CompetitiveLandscape
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.CompetitiveLandscape{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "CompetitiveLandscapes versioned and inserted"})
	case "operation":
		var req []models.OperationalEfficiency
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.OperationalEfficiency{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "OperationalEfficiencies versioned and inserted"})
	case "risk":
		var req []models.RiskManagement
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.RiskManagement{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "RiskManagements versioned and inserted"})
	case "additional":
		var req []models.AdditionalInfo
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.AdditionalInfo{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "AdditionalInfos versioned and inserted"})
	case "self":
		var req []models.SelfAssessment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.SelfAssessment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "SelfAssessments versioned and inserted"})
	case "attachements":
		var req []models.Attachment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.Attachment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = maxVersion + 1
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Attachments versioned and inserted"})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
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
	var db = values.GetDB()
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
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
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
	var db = values.GetDB()
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
	var db = values.GetDB()
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
	userID := claims.ID
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
