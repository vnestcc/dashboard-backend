package company

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

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
				if respondWithErrorIfNeeded(ctx, err, "finance") {
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

/*type MetricsStruct interface {

}*/

// CompanyMetrics godoc
// @Summary      Get company metric time series or snapshot
// @Description  Returns a specific company KPI or metric series (e.g. funds raised, revenue growth, runway, user growth, milestones, CAC/LTV, market share, KPIs, etc) by key.
// @Security     BearerAuth
// @Tags         company
// @Produce      json
// @Param        id    path   int     true  "Company ID"
// @Param        key   query  string  true  "Which metric to return" Enums(funds_raised, revenue_growth, revenue_breakdown, runway, user_growth, milestones, cac_ltv, market_share, kpis)
// @Success      200   {object}  map[string]any  "Success (company_id and metrics array/object)"
// @Failure      400   {object}  map[string]string  "Bad request (missing or invalid params)"
// @Failure      404   {object}  map[string]string  "Not found"
// @Failure      500   {object}  map[string]string  "Internal error"
// @Router       /company/metrics/{id} [get]
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
	case "funds_raised":
		var results []fundsRaisedMetric
		if err := db.Raw(`
		SELECT
			fin.cash_balance,
			fun.last_round,
			qua.quarter,
			qua.year,
			qua.date
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM finance
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) fin
		JOIN (
			SELECT DISTINCT ON (quarter_id) *
			FROM fund
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) fun ON fin.quarter_id = fun.quarter_id
		JOIN quarters qua ON fin.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID, company.ID).Scan(&results).Error; err != nil {
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
				"error":      err.Error(),
				"key":        key,
			})
			if logLevel == logrus.WarnLevel {
				entry.Warn("Funds raised query returned no data")
			} else {
				entry.Error("Failed to retrieve funds raised metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("Funds raised metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "revenue_growth":
		var results []revenueMetric
		if err := db.Raw(`
		SELECT
    	fin.quarterly_revenue,
	    fin.revenue_growth,
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
				entry.Warn("Revenue growth query returned no data")
			} else {
				entry.Error("Failed to retrieve revenue growth metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("Revenue growth metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "revenue_breakdown":
		var rows []revenueBreakdownRow
		err := db.Raw(`
		SELECT
			rb.product,
			rb.revenue,
			rb.percentage,
			qua.quarter,
			qua.year,
			qua.date
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM finance
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) AS fin
		JOIN revenue_breakdowns rb ON rb.financial_health_id = fin.id
		JOIN quarters qua ON fin.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID).Scan(&rows).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			}).Error("Failed to retrieve revenue breakdown metrics")
			return
		}
		grouped := make(map[string]*QuarterRevenueBreakdown)
		for _, row := range rows {
			key := fmt.Sprintf("%s-%d", row.Quarter, row.Year)
			if _, exists := grouped[key]; !exists {
				grouped[key] = &QuarterRevenueBreakdown{
					Quarter:    row.Quarter,
					Year:       row.Year,
					Date:       row.Date,
					Breakdowns: []Breakdown{},
				}
			}
			grouped[key].Breakdowns = append(grouped[key].Breakdowns, Breakdown{
				Product:    row.Product,
				Revenue:    row.Revenue,
				Percentage: row.Percentage,
			})
		}
		var results []QuarterRevenueBreakdown
		for _, v := range grouped {
			results = append(results, *v)
		}
		sort.Slice(results, func(i, j int) bool {
			if results[i].Year == results[j].Year {
				return results[i].Quarter < results[j].Quarter
			}
			return results[i].Year < results[j].Year
		})
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("Grouped revenue breakdowns retrieved")

		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "runway":
		var results []runwayMetric
		err := db.Raw(`
		SELECT
			fin.cash_balance,
			fin.burn_rate,
			fin.cash_runway,
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
		WHERE qua.company_id = ?
	`, company.ID, company.ID).Scan(&results).Error
		if err != nil {
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
				entry.Warn("Runway query returned no data")
			} else {
				entry.Error("Failed to retrieve runway metrics")
			}
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("Runway metrics retrieved successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "user_growth":
		var results []userGrowthMetric
		err := db.Raw(`
		SELECT
			pro.active_users,
			mark.total_customers,
			qua.quarter,
			qua.year,
			qua.date
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM product
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) AS pro
		JOIN (
			SELECT DISTINCT ON (quarter_id) *
			FROM market
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) AS mark ON pro.quarter_id = mark.quarter_id
		JOIN quarters qua ON pro.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID, company.ID).Scan(&results).Error
		if err != nil {
			status := "failure"
			reason := "db_error"
			logLevel := logrus.ErrorLevel
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
				reason = "record_not_found"
				logLevel = logrus.WarnLevel
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
			}
			auditLog.WithFields(logrus.Fields{
				"status":     status,
				"reason":     reason,
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			}).Log(logLevel, "Failed to retrieve user growth metrics")
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("User growth metrics retrieved")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "milestones":
		var rows []milestoneRow
		err := db.Raw(`
		SELECT
			pro.milestones_achieved,
			pro.roadmap,
			qua.quarter,
			qua.year
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM product
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) AS pro
		JOIN quarters qua ON pro.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID).Scan(&rows).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve milestones"})
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			}).Error("Failed to retrieve milestones data")
			return
		}
		milestonesAchieved := []string{}
		roadmap := []string{}
		for _, row := range rows {
			suffix := fmt.Sprintf("(%s %d)", row.Quarter, row.Year)
			if trimmed := strings.TrimSpace(row.MilestonesAchieved); trimmed != "" {
				milestonesAchieved = append(milestonesAchieved, fmt.Sprintf("%s %s", trimmed, suffix))
			}
			if trimmed := strings.TrimSpace(row.Roadmap); trimmed != "" {
				roadmap = append(roadmap, fmt.Sprintf("%s %s", trimmed, suffix))
			}
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(rows),
		}).Info("Milestones data retrieved")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"milestones": gin.H{
				"milestones_achieved": milestonesAchieved,
				"roadmap":             roadmap,
			},
		})
	case "cac_ltv":
		var results []cacLtvMetric
		err := db.Raw(`
		SELECT
			qua.date AS timestamp,
			qua.quarter,
			qua.year,
			eco.cac,
			eco.ltv,
			eco.ltv_ratio
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM economics
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) eco
		JOIN quarters qua ON eco.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID).Scan(&results).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve CAC/LTV data"})
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			}).Error("Failed to retrieve CAC/LTV data")
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("CAC/LTV data retrieved")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"metrics":    results,
		})
	case "market_share":
		var result struct {
			MarketShare    string `json:"market_share"`
			TotalCustomers string `json:"total_customers"`
		}
		err := db.Raw(`
		SELECT
			MAX(market_share) AS market_share,
			MAX(total_customers) AS total_customers
		FROM market
		WHERE company_id = ?
	`, company.ID).Scan(&result).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve market share"})
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"company_id": company.ID,
				"key":        key,
				"error":      err.Error(),
			}).Error("Failed to retrieve market share metrics")
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
		}).Info("Market share metrics retrieved")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"market_share": gin.H{
				"market_share":    result.MarketShare,
				"total_customers": result.TotalCustomers,
			},
		})
	case "kpis":
		var results []kpi
		err := db.Raw(`
		SELECT
			qua.date AS timestamp,
			qua.quarter,
			qua.year,
			pro.active_users,
			mar.conversion_rate,
			mar.churn_rate,
			fin.gross_margin
		FROM (
			SELECT DISTINCT ON (quarter_id) *
			FROM product
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) pro
		JOIN (
			SELECT DISTINCT ON (quarter_id) *
			FROM market
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) mar ON pro.quarter_id = mar.quarter_id
		JOIN (
			SELECT DISTINCT ON (quarter_id) *
			FROM finance
			WHERE company_id = ?
			ORDER BY quarter_id, version DESC
		) fin ON pro.quarter_id = fin.quarter_id
		JOIN quarters qua ON pro.quarter_id = qua.id
		WHERE qua.company_id = ?
	`, company.ID, company.ID, company.ID, company.ID).Scan(&results).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve KPI data"})
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"key":        key,
				"company_id": company.ID,
				"error":      err.Error(),
			}).Error("Failed to retrieve KPI metrics")
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status":     "success",
			"company_id": company.ID,
			"key":        key,
			"rows":       len(results),
		}).Info("KPI metrics retrieved")
		ctx.JSON(http.StatusOK, gin.H{
			"company_id": company.ID,
			"kpis":       results,
		})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid key '%s'", key)})
	}
}
