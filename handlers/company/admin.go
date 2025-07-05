package company

import (
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

// GetCompanyByIDAdmin godoc
// @Summary      Get company details (Admin)
// @Description  Returns the specified company's information, including selectable related data sets (admin only).
// @Tags         admin
// @Security     BearerAuth
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

func handleEditAdmin[T any](ctx *gin.Context, db *gorm.DB, quarterObj *models.Quarter, table string, auditLog *logrus.Entry) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "invalid_request_body",
			"table":  table,
		}).Warn("Invalid request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	var model T
	if err := db.Model(&model).Where("quarter_id = ? AND company_id = ?", quarterObj.ID, quarterObj.CompanyID).Order("version DESC").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"error":  "record_not_found",
				"table":  table,
			}).Warn("No existing record found to update")
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  err.Error(),
			"table":  table,
		}).Error("Database error while querying for record")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if err := db.Model(&model).Updates(req).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  err.Error(),
			"table":  table,
		}).Error("Failed to update record")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}
	v := reflect.ValueOf(model)
	getID := func() uint {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		idField := v.FieldByName("ID")
		if idField.IsValid() && idField.CanUint() {
			return uint(idField.Uint())
		}
		return 0
	}
	auditLog.WithFields(logrus.Fields{
		"status": "success",
		"table":  table,
		"id":     getID(),
	}).Info("Record updated successfully")
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s updated by admin", table),
		"id":      getID(),
	})
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
// TEST: need testing
func EditCompanyByID(ctx *gin.Context) {
	db := values.GetDB()
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":         ctx.ClientIP(),
		"type":       "audit",
		"event":      "edit_company",
		"company_id": companyID,
	})
	var company models.Company
	if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "company_not_found",
		}).Warn("Company not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
		return
	}
	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "multiple_data_params",
		}).Warn("More than one 'data' parameter supplied")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Only one 'data' query parameter allowed"})
		return
	}
	data := ""
	if len(dataVals) == 1 {
		data = dataVals[0]
	}
	quarter := ctx.Query("quarter")
	yearStr := ctx.Query("year")
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "invalid_year",
			"value":  yearStr,
		}).Warn("Invalid year")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
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
		"product":       "ProductDevelopment",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"error":  "invalid_data_param",
				"value":  data,
			}).Warn("Invalid 'data' query parameter")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	if data == "" || data == "info" {
		var req struct {
			Name         *string `json:"name"`
			ContactName  *string `json:"contact_name"`
			ContactEmail *string `json:"contact_email"`
		}
		infoLog := auditLog.WithField("table", "companies")
		if err := ctx.ShouldBindJSON(&req); err != nil {
			infoLog.WithFields(logrus.Fields{
				"status": "failure",
				"error":  "invalid_request_body",
				"detail": err.Error(),
			}).Warn("Failed to parse request body")
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
			infoLog.WithFields(logrus.Fields{
				"status": "failure",
				"error":  err.Error(),
			}).Error("Failed to update company info")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
			return
		}
		infoLog.WithFields(logrus.Fields{
			"status": "success",
			"id":     company.ID,
		}).Info("Company info updated successfully")
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Company updated successfully",
			"company": company,
		})
		return
	}
	cacheKey := fmt.Sprintf("%d_%s_%d", companyID, quarter, year)
	var quarterObj models.Quarter
	if val, ok := QuarterCache.Get(cacheKey); ok {
		quarterObj = val
	} else {
		if err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).First(&quarterObj).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"status":  "failure",
				"error":   "quarter_not_found",
				"quarter": quarter,
				"year":    year,
			}).Warn("Quarter not found for company")
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
			return
		}
		QuarterCache.Set(cacheKey, quarterObj)
	}
	preloadField := allowedData[data]
	if preloadField == "" {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "no_editable_data_specified",
		}).Warn("Preload field is empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No editable data specified"})
		return
	}
	sectionLog := auditLog.WithFields(logrus.Fields{
		"quarter": quarter,
		"year":    year,
		"table":   preloadField,
	})
	switch data {
	case "finance":
		handleEditAdmin[*models.FinancialHealth](ctx, db, &quarterObj, preloadField, sectionLog)
	case "market":
		handleEditAdmin[*models.MarketTraction](ctx, db, &quarterObj, preloadField, sectionLog)
	case "uniteconomics":
		handleEditAdmin[*models.UnitEconomics](ctx, db, &quarterObj, preloadField, sectionLog)
	case "teamperf":
		handleEditAdmin[*models.TeamPerformance](ctx, db, &quarterObj, preloadField, sectionLog)
	case "product":
		handleEditAdmin[*models.ProductDevelopment](ctx, db, &quarterObj, preloadField, sectionLog)
	case "fund":
		handleEditAdmin[*models.FundraisingStatus](ctx, db, &quarterObj, preloadField, sectionLog)
	case "competitive":
		handleEditAdmin[*models.CompetitiveLandscape](ctx, db, &quarterObj, preloadField, sectionLog)
	case "operation":
		handleEditAdmin[*models.OperationalEfficiency](ctx, db, &quarterObj, preloadField, sectionLog)
	case "risk":
		handleEditAdmin[*models.RiskManagement](ctx, db, &quarterObj, preloadField, sectionLog)
	case "additional":
		handleEditAdmin[*models.AdditionalInfo](ctx, db, &quarterObj, preloadField, sectionLog)
	case "self":
		handleEditAdmin[*models.SelfAssessment](ctx, db, &quarterObj, preloadField, sectionLog)
	default:
		sectionLog.WithField("status", "failure").Warn("Unexpected data type after validation")
		// attachments comes under upload
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
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
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "delete_company_by_id",
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
	tx := db.Delete(&models.Company{}, companyID)
	if tx.Error != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "company_delete_failed",
			"company_id": companyID,
			"error":      tx.Error.Error(),
		}).Error("Failed to delete company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}
	if tx.RowsAffected == 0 {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "company_not_exist",
			"company_id": companyID,
		}).Warn("Company does not exist")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Company does not exist"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("startup_id = ?", companyID).
		Update("startup_id", nil).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "unlink_users_failed",
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to clear company from users")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from users"})
		return
	}
	StartupCache.Delete(companyID)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": companyID,
	}).Info("Company deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

// AllowQuarterByID godoc
// @Summary      Set next allowed quarter/year for a company
// @Description  Allows a moderator to define the next quarter and year that a company is allowed to create. This updates the `planned_quarter` and `planned_year` fields for the company.
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int              true  "Company ID"
// @Param        body body      nextQuarter       true  "Quarter and Year to allow"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/quarters/{id}/new [post]
func AllowQuarterByID(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "allow_quarter_by_id",
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
	if val, found := StartupCache.Get(companyID); found {
		company = val
	} else {
		if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"status": "failure",
					"error":  "company_not_found",
				}).Warn("Company not found")
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
				return
			} else {
				auditLog.WithFields(logrus.Fields{
					"status": "failure",
					"error":  "internal_server_error",
				}).Error("Error while retriving company")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retriving company"})
				return
			}
		}
		StartupCache.Set(companyID, company)
	}
	var request nextQuarter
	if err := ctx.ShouldBindJSON(&request); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_request_body",
			"company_id": companyID,
			"error":      err.Error(),
		}).Warn("Failed to bind request JSON")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	validQuarters := map[string]bool{"Q1": true, "Q2": true, "Q3": true, "Q4": true}
	if !validQuarters[request.NextQuarter] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quarter. Must be one of Q1, Q2, Q3, Q4"})
		return
	}
	if err := db.Model(&company).Updates(map[string]any{
		"planned_quarter": &request.NextQuarter,
		"planned_year":    &request.NextYear,
	}).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "db_update_failed",
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to update company with planned quarter/year")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}
	StartupCache.Set(companyID, company)
	auditLog.WithFields(logrus.Fields{
		"status":          "success",
		"company_id":      companyID,
		"planned_quarter": request.NextQuarter,
		"planned_year":    request.NextYear,
	}).Info("Successfully updated company's next quarter and year")
	ctx.JSON(http.StatusOK, gin.H{"message": "Company updated with next quarter/year"})
}

// RemoveQuarterByID godoc
// @Summary      Remove planned quarter and year for a company
// @Description  Allows a moderator to unset (nullify) the planned_quarter and planned_year fields for a company.
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
// @Router       /manage/company/quarters/{id}/remove [delete]
func RemoveQuarterByID(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "allow_quarter_by_id",
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
	if val, found := StartupCache.Get(companyID); found {
		company = val
	} else {
		if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"status": "failure",
					"error":  "company_not_found",
				}).Warn("Company not found")
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
				return
			} else {
				auditLog.WithFields(logrus.Fields{
					"status": "failure",
					"error":  "internal_server_error",
				}).Error("Error while retriving company")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retriving company"})
				return
			}
		}
		StartupCache.Set(companyID, company)
	}
	if err := db.Model(&company).Updates(map[string]any{
		"planned_quarter": nil,
		"planned_year":    nil,
	}).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "db_update_failed",
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to update company with planned quarter/year")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}
	StartupCache.Set(companyID, company)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": companyID,
	}).Info("Successfully removed company's next quarter and year")
	ctx.JSON(http.StatusOK, gin.H{"message": "Company updated with next quarter/year as nil"})
}

// AllowQuarter godoc
// @Summary      Set next allowed quarter/year for all companies
// @Description  Allows a moderator to define the next quarter and year that all companies are allowed to create. This updates the `planned_quarter` and `planned_year` fields for all companies.
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body      nextQuarter  true  "Quarter and Year to allow"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/quarters/new [post]
func AllowQuarter(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "allow_quarter_all",
	})
	var request nextQuarter
	if err := ctx.ShouldBindJSON(&request); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "invalid_request_body",
			"error":  err.Error(),
		}).Warn("Failed to bind request JSON")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	validQuarters := map[string]bool{"Q1": true, "Q2": true, "Q3": true, "Q4": true}
	if !validQuarters[request.NextQuarter] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quarter. Must be one of Q1, Q2, Q3, Q4"})
		return
	}
	if err := db.Model(&models.Company{}).Updates(map[string]any{
		"planned_quarter": &request.NextQuarter,
		"planned_year":    &request.NextYear,
	}).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_update_failed",
			"error":  err.Error(),
		}).Error("Failed to update companies with planned quarter/year")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update companies"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status":          "success",
		"planned_quarter": request.NextQuarter,
		"planned_year":    request.NextYear,
	}).Info("Successfully updated all companies' next quarter and year")
	ctx.JSON(http.StatusOK, gin.H{"message": "All companies updated with next quarter/year"})
}

// RemoveQuarter godoc
// @Summary      Remove planned quarter and year for all companies
// @Description  Allows a moderator to unset (nullify) the planned_quarter and planned_year fields for all companies.
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/quarters/remove [delete]
func RemoveQuarter(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "remove_quarter_all",
	})
	if err := db.Model(&models.Company{}).Updates(map[string]any{
		"planned_quarter": nil,
		"planned_year":    nil,
	}).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_update_failed",
			"error":  err.Error(),
		}).Error("Failed to update companies with planned quarter/year as nil")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update companies"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status": "success",
	}).Info("Successfully removed planned quarter and year for all companies")
	ctx.JSON(http.StatusOK, gin.H{"message": "All companies updated with next quarter/year as nil"})
}
