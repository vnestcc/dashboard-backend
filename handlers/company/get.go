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
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
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
		var results []*models.FinancialHealth
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.FinancialHealth](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "finance") {
			return
		}
		for _, fh := range results {
			err := db.Model(fh).Association("RevenueBreakdowns").Find(&fh.RevenueBreakdowns)
			if err != nil {
				if respondWithErrorIfNeeded(ctx, err, "finance") {
					return
				}
			}
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "FinancialHealths"}).Info("Fetched FinancialHealth data")
	case "market":
		var results []*models.MarketTraction
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.MarketTraction](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "market") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "MarketTractions"}).Info("Fetched MarketTraction data")
	case "economics":
		var results []*models.UnitEconomics
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.UnitEconomics](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "economics") {
			return
		}
		for _, ue := range results {
			err := db.Model(ue).Association("MarketingBreakdowns").Find(&ue.MarketingBreakdowns)
			if err != nil {
				if respondWithErrorIfNeeded(ctx, err, "economics") {
					return
				}
			}
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "UnitEconomics"}).Info("Fetched UnitEconomics data")
	case "teamperf":
		var results []*models.TeamPerformance
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.TeamPerformance](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "teamperf") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "TeamPerformances"}).Info("Fetched TeamPerformance data")
	case "fund":
		var results []*models.FundraisingStatus
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.FundraisingStatus](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "fund") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "FundraisingStatuses"}).Info("Fetched FundraisingStatus data")
	case "competitive":
		var results []*models.CompetitiveLandscape
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.CompetitiveLandscape](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "competitve") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "CompetitiveLandscapes"}).Info("Fetched CompetitiveLandscape data")
	case "product":
		var results []*models.ProductDevelopment
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.ProductDevelopment](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "product") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "ProductDevelopment"}).Info("Fetched ProductDevelopment data")
	case "operational":
		var results []*models.OperationalEfficiency
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.OperationalEfficiency](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "operational") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "OperationalEfficiencies"}).Info("Fetched OperationalEfficiency data")
	case "risk":
		var results []*models.RiskManagement
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.RiskManagement](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "risk") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "RiskManagements"}).Info("Fetched RiskManagement data")
	case "additional":
		var results []*models.AdditionalInfo
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.AdditionalInfo](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "additional") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "AdditionalInfos"}).Info("Fetched AdditionalInfo data")
	case "assessment":
		var results []*models.SelfAssessment
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.SelfAssessment](db, companyID, quarter, year, tableName)
		}
		if respondWithErrorIfNeeded(ctx, err, "assessment") {
			return
		}
		filterAndRespond(ctx, results, quarterID, fullAccess)
		auditLog.WithFields(logrus.Fields{"status": "success", "data_type": "SelfAssessments"}).Info("Fetched SelfAssessment data")
	case "attachment":
		var results []*models.Attachment
		if found {
			err = db.Where("quarter_id = ? AND company_id = ?", quarterObj.ID, companyID).Find(&results).Error
			quarterID = quarterObj.ID
		} else {
			results, quarterID, err = queryByQuarterID[*models.Attachment](db, companyID, quarter, year, tableName)
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
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements, product)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/{id} [get]
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
		"product":       "product",
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

// CompanyMetrics godoc
// @Summary     Retrieve company KPI or metric series
// @Description Returns either a time series or snapshot of a specified company KPI or metric (such as funds raised, revenue growth, runway, user growth, milestones, CAC/LTV, market share, or other KPIs) based on the provided key.
// @Security    BearerAuth
// @Tags        company
// @Produce     json
// @Param       id   path   int    true  "Company ID"
// @Param       key  query  string true  "Metric key" Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements, product)
// @Success     200  {object} map[string]any          "Success (company_id and metrics array/object)"
// @Failure     400  {object} map[string]string       "Bad request (missing or invalid params)"
// @Failure     404  {object} map[string]string       "Not found"
// @Failure     500  {object} map[string]string       "Internal server error"
// @Router      /company/metrics/{id} [get]
func CompanyMetrics(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "get_company_metrics_by_id",
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
	keys := ctx.Request.URL.Query()["key"]
	if len(keys) != 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "You must provide exactly one 'key' query parameter"})
		return
	}
	key := keys[0]
	switch key {
	case "finance":
		var results []financeMetric
		if err := db.Raw(`
		SELECT
    	fin.quarterly_revenue,
	    fin.revenue_growth,
			fin.gross_margin,
			fin.net_margin,
  	  qua.quarter,
    	qua.year,
	    qua.date
		FROM (
  	  SELECT DISTINCT ON (quarter_id) *
    	FROM finance
	    WHERE company_id = ?
  	  ORDER BY quarter_id, version DESC
		) AS fin
		JOIN quarters qua ON fin.quarter_id = qua.id
		WHERE qua.company_id = ?;
	`, company.ID, company.ID).Scan(&results).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Finance query returned no data")
			} else {
				entry.Error("Failed to retrieve finance metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("Finance metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})

	case "market":
		var metrics []marketMetric
		if err := db.Raw(`
        SELECT 
            m.total_customers, 
            m.customer_growth, 
            m.conversion_rate, 
            m.retention_rate, 
            m.churn_rate, 
            q.quarter, 
            q.year, 
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM market
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) m
        JOIN quarters q ON m.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Market query returned no data")
			} else {
				entry.Error("Failed to retrieve market metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Market metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "economics":
		var metrics []economicsMetric
		if err := db.Raw(`
        SELECT 
            e.cac, 
            e.cac_payback, 
            e.arpu, 
            e.ltv, 
            q.quarter, 
            q.year, 
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM economics
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) e
        JOIN quarters q ON e.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Economics query returned no data")
			} else {
				entry.Error("Failed to retrieve economics metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Economics metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "product":
		var metrics []productMetric
		if err := db.Raw(`
        SELECT 
            p.active_users, 
            p.engagement_metrics, 
            p.milestones_achieved, 
            p.milestones_missed, 
            p.roadmap, 
            p.technical_challenges, 
            p.product_bottlenecks, 
            q.quarter, 
            q.year, 
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM product
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) p
        JOIN quarters q ON p.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Product query returned no data")
			} else {
				entry.Error("Failed to retrieve product metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Product metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "teamperf":
		var metrics []teamperfMetric
		if err := db.Raw(`
				SELECT 
						t.team_strengths,
						t.development_initiatives,
						t.team_size,
						t.new_hires,
						t.turnover,
						t.vacant_positions,
						t.leadership_alignment,
						t.skill_gaps,
						q.quarter,
						q.year,
						q.date
				FROM (
						SELECT DISTINCT ON (quarter_id) *
						FROM teamperf
						WHERE company_id = ?
						ORDER BY quarter_id, version DESC
				) t
				JOIN quarters q ON t.quarter_id = q.id
				WHERE q.company_id = ?
				ORDER BY q.year, q.quarter ASC
		`, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Teamperf query returned no data")
			} else {
				entry.Error("Failed to retrieve teamperf metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Teamperf metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "fund":
		var metrics []fundMetric
		if err := db.Raw(`
        SELECT 
            f.last_round,
            f.target_amount,
            f.valuation_expectations,
            q.quarter,
            q.year,
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM fund
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) f
        JOIN quarters q ON f.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Fund query returned no data")
			} else {
				entry.Error("Failed to retrieve fund metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Fund metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "operational":
		var metrics []operationalMetric
		if err := db.Raw(`
        SELECT 
            o.infrastructure_capacity,
            o.operational_bottlenecks,
            o.optimization_areas,
            o.scaling_plans,
            q.quarter,
            q.year,
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM operational
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) o
        JOIN quarters q ON o.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Operational query returned no data")
			} else {
				entry.Error("Failed to retrieve operational metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Operational metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "risk":
		var metrics []riskMetric
		if err := db.Raw(`
        SELECT 
            r.strategic_risks,
            r.operational_risks,
            r.financial_risks,
            r.legal_risks,
            r.regulatory_risks,
            r.mitigation_plans,
            q.quarter,
            q.year,
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM risk
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) r
        JOIN quarters q ON r.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Risk query returned no data")
			} else {
				entry.Error("Failed to retrieve risk metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Risk metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "additional":
		var metrics []additionalMetric
		if err := db.Raw(`
        SELECT 
            a.customer_feedback,
            a.market_trends,
            a.regulatory_changes,
            a.noteworthy_events,
            q.quarter,
            q.year,
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM additional
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) a
        JOIN quarters q ON a.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Additional query returned no data")
			} else {
				entry.Error("Failed to retrieve additional metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Additional metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	case "assessment":
		var metrics []assessmentMetric
		if err := db.Raw(`
        SELECT 
            a.assessment_text,
            a.assessment_score,
            q.quarter,
            q.year,
            q.date
        FROM (
            SELECT DISTINCT ON (quarter_id) *
            FROM assessment
            WHERE company_id = ?
            ORDER BY quarter_id, version DESC
        ) a
        JOIN quarters q ON a.quarter_id = q.id
        WHERE q.company_id = ?
        ORDER BY q.year, q.quarter ASC
    `, company.ID, company.ID).Scan(&metrics).Error; err != nil {
			logLevel := logrus.ErrorLevel
			reason := "db_error"
			status := "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server issue"})
			}
			entry := auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Assessment query returned no data")
			} else {
				entry.Error("Failed to retrieve assessment metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(metrics),
		}).Info("Assessment metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    metrics,
		})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid key '%s'", key)})
	}
}
