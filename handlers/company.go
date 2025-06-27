package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
	"github.com/vnestcc/dashboard/utils/values"
	"gorm.io/gorm"
)

var StartupCache = cacher.NewCacher[uint, models.Company](&cacher.NewCacherOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
	CleanerMode:   cacher.CleaningCentral,
})

type Claims = middleware.Claims

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
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "user_company",
	})
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized access: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		auditLog.WithField("status", "failure").Warn("Invalid claims format")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
			"error":   err.Error(),
		}).Error("Failed to fetch user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartUp == nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
		}).Warn("User does not belong to any company")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to any company"})
		return
	}
	startup := user.StartUp
	if _, ok := StartupCache.Get(startup.ID); !ok {
		StartupCache.Set(startup.ID, *startup)
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"user_id":    claims.ID,
		"company_id": startup.ID,
	}).Info("Fetched user's company")
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
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "list_quarter",
	})
	idStr := ctx.Param("id")
	companyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_company_id",
			"company_id": idStr,
		}).Warn("Invalid company ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var company models.Company
	if cached, found := StartupCache.Get(uint(companyID)); found {
		company = cached
	} else {
		if err := db.First(&company, uint(companyID)).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"reason":     "company_not_found",
				"company_id": companyID,
				"error":      err.Error(),
			}).Warn("Company not found")
			ctx.Set("message", err.Error())
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch company"})
			return
		}
		StartupCache.Set(company.ID, company)
	}
	if err := db.Model(&company).Association("Quarters").Find(&company.Quarters); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "fetch_quarters_failed",
			"company_id": company.ID,
			"error":      err.Error(),
		}).Error("Failed to fetch quarters")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quarters"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": company.ID,
		"quarters":   len(company.Quarters),
	}).Info("Fetched company quarters")
	ctx.JSON(http.StatusOK, company.Quarters)
}

// ListCompanyAdmin godoc
// @Summary      List all companies
// @Description  Retrieves a list of all companies available in the system
// @Tags         company
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /manage/company/list [get]
func ListCompanyAdmin(ctx *gin.Context) {
	// placeholder func
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
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "list_company",
	})
	var companies []models.Company
	result := make(map[uint]string)
	if err := db.Find(&companies).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  err.Error(),
		}).Error("Failed to retrieve the company list")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the company list"})
		return
	}
	for i := range companies {
		StartupCache.Set(companies[i].ID, companies[i])
		result[companies[i].ID] = companies[i].Name
	}
	auditLog.WithFields(logrus.Fields{
		"status":        "success",
		"company_count": len(result),
	}).Info("Fetched company list")
	ctx.JSON(http.StatusOK, result)
}

var QuarterCache = cacher.NewCacher[string, models.Quarter](&cacher.NewCacherOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
	CleanerMode:   cacher.CleaningLocal,
})

func extractQuarterID(item any) uint {
	val := reflect.ValueOf(item)
	field := val.FieldByName("QuarterID")
	if field.IsValid() && field.CanUint() {
		return uint(field.Uint())
	}
	return 0
}

func respondWithErrorIfNeeded(ctx *gin.Context, err error, label string) bool {
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "data_section_error",
		"label": label,
	})
	if err == nil {
		return false
	}
	ctx.Set("message", err.Error())
	if err.Error() == "no elements exist" || errors.Is(err, gorm.ErrRecordNotFound) {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "not_found",
			"error":  err.Error(),
		}).Warn("No data found for label")
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No %s found for the given quarter/year", label)})
	} else {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_error",
			"error":  err.Error(),
		}).Error("Could not load data for label")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not load %s", label)})
	}
	return true
}

func filterAndRespond[T any](ctx *gin.Context, data []T, quarterID uint, fullAccess bool) {
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":         ctx.ClientIP(),
		"type":       "audit",
		"event":      "filter_and_respond",
		"quarter_id": quarterID,
	})
	filtered := []map[string]any{}
	for _, item := range data {
		if f, ok := any(item).(interface {
			VisibilityFilter(bool) map[string]any
		}); ok {
			filtered = append(filtered, f.VisibilityFilter(fullAccess))
		}
	}
	auditLog.WithFields(logrus.Fields{
		"status":   "success",
		"filtered": len(filtered),
	}).Info("Responding with filtered data")
	ctx.JSON(http.StatusOK, gin.H{
		"quarter_id": quarterID,
		"data":       filtered,
	})
}

func queryByQuarterID[T any](db *gorm.DB, companyID uint, quarter string, year uint, tableName string) ([]T, uint, error) {
	var results []T
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE company_id = ?
		AND quarter_id = (
			SELECT id FROM quarters
			WHERE company_id = ? AND quarter = ? AND year = ?
			LIMIT 1
		)
	`, tableName)
	err := db.Raw(query, companyID, companyID, quarter, year).Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}
	if len(results) == 0 {
		return results, 0, errors.New("no elements exist")
	}
	quarterID := extractQuarterID(results[0])
	if quarterID == 0 {
		return results, 0, errors.New("QuarterID not found")
	}
	return results, quarterID, nil
}

func handleDataSection(ctx *gin.Context, db *gorm.DB, companyID uint, quarter string, year uint, tableName string, fullAccess bool) {
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":         ctx.ClientIP(),
		"type":       "audit",
		"event":      "handle_data_section",
		"company_id": companyID,
		"quarter":    quarter,
		"year":       year,
		"table":      tableName,
	})
	cacheKey := fmt.Sprintf("%d_%s_%d", companyID, quarter, year)
	quarterObj, found := QuarterCache.Get(cacheKey)
	var quarterID uint
	var err error
	switch tableName {
	case "finance":
		var results []models.FinancialHealth
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.FinancialHealth](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "FinancialHealths") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "FinancialHealths"}).Info("Fetched FinancialHealth data")
	case "market":
		var results []models.MarketTraction
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.MarketTraction](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "MarketTractions") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "MarketTractions"}).Info("Fetched MarketTraction data")
	case "economics":
		var results []models.UnitEconomics
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.UnitEconomics](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "UnitEconomics") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "UnitEconomics"}).Info("Fetched UnitEconomics data")
	case "teamperf":
		var results []models.TeamPerformance
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.TeamPerformance](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "TeamPerformances") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "TeamPerformances"}).Info("Fetched TeamPerformance data")
	case "fund":
		var results []models.FundraisingStatus
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.FundraisingStatus](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "FundraisingStatuses") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "FundraisingStatuses"}).Info("Fetched FundraisingStatus data")
	case "competitive":
		var results []models.CompetitiveLandscape
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.CompetitiveLandscape](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "CompetitiveLandscapes") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "CompetitiveLandscapes"}).Info("Fetched CompetitiveLandscape data")
	case "operational":
		var results []models.OperationalEfficiency
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.OperationalEfficiency](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "OperationalEfficiencies") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "OperationalEfficiencies"}).Info("Fetched OperationalEfficiency data")
	case "risk":
		var results []models.RiskManagement
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.RiskManagement](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "RiskManagements") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "RiskManagements"}).Info("Fetched RiskManagement data")
	case "additional":
		var results []models.AdditionalInfo
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.AdditionalInfo](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "AdditionalInfos") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "AdditionalInfos"}).Info("Fetched AdditionalInfo data")
	case "assessment":
		var results []models.SelfAssessment
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.SelfAssessment](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "SelfAssessments") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "SelfAssessments"}).Info("Fetched SelfAssessment data")
	case "attachment":
		var results []models.Attachment
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[models.Attachment](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "Attachments") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "Attachments"}).Info("Fetched Attachment data")
	default:
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "unsupported_data_type",
			"table":  tableName,
		}).Warn("Unsupported data type")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported data type"})
	}
}

// GetCompanyByID godoc
// @Summary      Get company details
// @Description  Returns the current user's company information, including selectable related data sets
// @Tags         company
// @Produce      json
// @Param        id       path   int     true  "Company ID"
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/{id} [get]
// TEST: testing
func GetCompanyByID(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "get_company_by_id",
	})
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_company_id",
			"company_id": idStr,
		}).Warn("Invalid company ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)
	var company models.Company
	if value, ok := StartupCache.Get(companyID); ok {
		company = value
	} else {
		if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"status":     "failure",
					"reason":     "company_not_found",
					"company_id": companyID,
				}).Warn("Company not found")
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
			} else {
				auditLog.WithFields(logrus.Fields{
					"status":     "failure",
					"reason":     "db_error",
					"company_id": companyID,
					"error":      err.Error(),
				}).Error("Failed to retrieve company")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company"})
			}
			return
		}
		StartupCache.Set(companyID, company)
	}
	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "too_many_data_query_params",
		}).Warn("Only one 'data' query parameter allowed")
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
		"finance":       "finance",
		"market":        "market",
		"uniteconomics": "economics",
		"teamperf":      "teamperf",
		"fund":          "fund",
		"competitive":   "competitive",
		"operation":     "operational",
		"risk":          "risk",
		"additional":    "additional",
		"self":          "assessment",
		"attachements":  "attachement",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"reason": "invalid_data_param",
				"data":   data,
			}).Warn("Invalid data query parameter")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	if data == "" || data == "info" {
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"info_only":  true,
		}).Info("Fetched company info")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id":            company.ID,
			"company_name":          company.Name,
			"company_contact_name":  company.ContactName,
			"company_contact_email": company.ContactEmail,
		})
		return
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "invalid_year",
			"year":   yearStr,
		}).Warn("Invalid year")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "unauthorized",
		}).Warn("Unauthorized")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "invalid_claims",
		}).Warn("Invalid claims format")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var access sql.NullInt64
	err = db.Raw(`
		SELECT CASE WHEN ? = 'admin' OR startup_id = ? THEN 1 ELSE 0 END 
		FROM users WHERE id = ?
	`, claims.Role, companyID, claims.ID).Scan(&access).Error
	if err != nil || !access.Valid {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "not_authorized_or_not_found",
			"company_id": companyID,
			"user_id":    claims.ID,
		}).Warn("User not authorized or not found")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized or not found"})
		return
	}
	fullAccess := access.Int64 == 1
	table := allowedData[data]
	auditLog.WithFields(logrus.Fields{
		"status":      "success",
		"company_id":  companyID,
		"data":        data,
		"quarter":     quarter,
		"year":        year,
		"full_access": fullAccess,
		"user_id":     claims.ID,
	}).Info("Fetching company data section")
	handleDataSection(ctx, db, companyID, quarter, year, table, fullAccess)
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
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "create_company",
	})

	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok || claims.Role == "admin" {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"role":   claims.Role,
		}).Warn("Admins cannot create companies")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot create companies"})
		return
	}
	var req struct {
		Name         string `json:"name" binding:"required"`
		ContactName  string `json:"contact_name" binding:"required"`
		ContactEmail string `json:"contact_email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"reason":  "invalid_input",
			"details": err.Error(),
		}).Warn("Invalid input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	var user models.User
	if err := db.First(&user, claims.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
			"error":   err.Error(),
		}).Error("Failed to fetch user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartupID != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": user.ID,
			"reason":  "already_in_company",
		}).Warn("User already belongs to a company")
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
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"reason": "duplicate_email",
				"email":  req.ContactEmail,
				"error":  err.Error(),
			}).Warn("Company with this contact email already exists")
			ctx.JSON(http.StatusConflict, gin.H{"error": "Company with this contact email already exists"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_error",
			"error":  err.Error(),
		}).Error("Failed to create company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}
	if err := db.Model(&user).Update("startup_id", newCompany.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"user_id":    user.ID,
			"company_id": newCompany.ID,
			"error":      err.Error(),
		}).Error("Failed to link company to user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link company to user"})
		return
	}
	StartupCache.Set(newCompany.ID, newCompany)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": newCompany.ID,
		"user_id":    user.ID,
	}).Info("Company created and linked to user")
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Company created successfully",
		"company_id": newCompany.ID,
	})
}

// GetCompanyByIDAdmin godoc
// @Summary      Get company details (Admin)
// @Description  Returns the specified company's information, including selectable related data sets (admin only).
// @Tags         admin
// @Produce      json
// @Param        id       path   int     true  "Company ID"
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/{id} [get]
func GetCompanyByIDAdmin(ctx *gin.Context) {
	// dead function just a place holder for docs
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
	claims, ok := claimsVal.(*Claims)
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
	claims, ok := claimsVal.(*Claims)
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
	if err := db.Model(&models.User{}).
		Where("startup_id = ?", companyID).
		Update("startup_id", nil).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from users"})
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
	claims, ok := claimsVal.(*Claims)
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
