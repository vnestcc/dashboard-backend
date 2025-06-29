package company

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	"github.com/vnestcc/dashboard/utils/values"
	"gorm.io/gorm"
)

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
		if respondWithErrorIfNeeded(ctx, err, "finance") {
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
		if respondWithErrorIfNeeded(ctx, err, "market") {
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
		if respondWithErrorIfNeeded(ctx, err, "economics") {
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
		if respondWithErrorIfNeeded(ctx, err, "teamperf") {
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
		if respondWithErrorIfNeeded(ctx, err, "fund") {
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
		if respondWithErrorIfNeeded(ctx, err, "competitve") {
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
		if respondWithErrorIfNeeded(ctx, err, "operational") {
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
		if respondWithErrorIfNeeded(ctx, err, "risk") {
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
		if respondWithErrorIfNeeded(ctx, err, "additional") {
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
		if respondWithErrorIfNeeded(ctx, err, "assessment") {
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
		if respondWithErrorIfNeeded(ctx, err, "attachment") {
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
// @Security     BearerAuth
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
